package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/joho/godotenv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/app"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/client"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/content"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/telemetry"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/theme"
)

const (
	defaultHost      = "0.0.0.0"
	defaultPort      = "2222"
	idleTimeout      = 10 * time.Minute
	maxSessionsPerIP = 5
)

func main() {
	// Load .env file (ignore error if not found)
	_ = godotenv.Load()

	// Initialize logger
	logger := telemetry.NewLogger("tui-server")

	// Initialize analytics
	analytics := telemetry.NewAnalytics(logger)
	defer analytics.Close()

	// Configuration from environment
	host := getEnv("SSH_HOST", defaultHost)
	port := getEnv("SSH_PORT", defaultPort)
	aiGatewayURL := getEnv("AI_GATEWAY_URL", "http://localhost:3001")
	contentPath := getEnv("CONTENT_PATH", getContentPath())

	logger.Info("Starting SSH server", telemetry.Ctx(
		"host", host,
		"port", port,
		"aiGateway", aiGatewayURL,
		"contentPath", contentPath,
	))

	// Track server start
	analytics.TrackServerStart(host, port)

	// Load content
	contentLoader := content.NewLoader(contentPath)

	resume, err := contentLoader.LoadResume()
	if err != nil {
		logger.Error("Failed to load resume", telemetry.Ctx("error", err.Error()))
		os.Exit(1)
	}
	logger.Debug("Resume loaded successfully")

	projects, err := contentLoader.LoadProjects()
	if err != nil {
		logger.Error("Failed to load projects", telemetry.Ctx("error", err.Error()))
		os.Exit(1)
	}
	logger.Debug("Projects loaded", telemetry.Ctx("count", len(projects.Projects)))

	bio, err := contentLoader.LoadBio()
	if err != nil {
		logger.Error("Failed to load bio", telemetry.Ctx("error", err.Error()))
		os.Exit(1)
	}
	logger.Debug("Bio loaded successfully")

	// Create AI client
	aiClient := client.NewAIClient(aiGatewayURL)

	// Check AI gateway health (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := aiClient.Health(ctx); err != nil {
			logger.Warn("AI gateway not available", telemetry.Ctx("error", err.Error()))
		} else {
			logger.Info("AI gateway connected")
		}
	}()

	// Session counter for rate limiting
	sessionCounter := NewSessionCounter(maxSessionsPerIP)

	// Create SSH server
	s, err := wish.NewServer(
		wish.WithAddress(host+":"+port),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithIdleTimeout(idleTimeout),
		wish.WithMiddleware(
			// Bubble Tea middleware
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				sessionStart := time.Now()

				// Extract comprehensive session info (PII-safe)
				sessionInfo := telemetry.ExtractSessionInfo(s)

				// Use hashed session ID for privacy
				sessionID := sessionInfo.SessionHash

				// Get terminal size
				pty, _, active := s.Pty()
				if !active {
					logger.Warn("No PTY for session", telemetry.Ctx(
						"session_hash", sessionID,
						"user_hash", sessionInfo.UserHash,
					))
					return nil, nil
				}

				width := pty.Window.Width
				height := pty.Window.Height
				if width == 0 {
					width = 80
				}
				if height == 0 {
					height = 24
				}

				// Update session info with validated dimensions
				sessionInfo.TerminalWidth = width
				sessionInfo.TerminalHeight = height

				// Log comprehensive session data (all PII-safe)
				logger.Info("Session connected", sessionInfo.ToMap())

				// Track session with full info
				analytics.TrackSessionConnectedWithInfo(sessionInfo)

				// Create renderer tied to SSH session for proper color support
				renderer := bubbletea.MakeRenderer(s)

				// Create session-specific theme manager with the renderer
				themeManager := theme.NewManager(width, height, renderer)

				// Create model with analytics
				model := app.NewModel(app.Config{
					ThemeManager: themeManager,
					Resume:       resume,
					Projects:     projects,
					Bio:          bio,
					AIClient:     aiClient,
					SessionID:    sessionID,
					Width:        width,
					Height:       height,
					Analytics:    analytics,
				})

				// Track disconnect on session end
				go func() {
					<-s.Context().Done()
					duration := time.Since(sessionStart).Milliseconds()
					logger.Info("Session disconnected", telemetry.Ctx(
						"session_hash", sessionID,
						"user_hash", sessionInfo.UserHash,
						"duration_ms", duration,
						"terminal", sessionInfo.Terminal,
					))
					analytics.TrackSessionDisconnected(sessionID, duration)
				}()

				return model, []tea.ProgramOption{
					tea.WithAltScreen(),
				}
			}),
			// Active terminal middleware (ensures PTY)
			activeterm.Middleware(),
			// Session rate limiting
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					addr := s.RemoteAddr().String()
					if !sessionCounter.Acquire(addr) {
						// Hash IP for logging (PII-safe)
						logger.Warn("Rate limited connection", telemetry.Ctx(
							"ip_hash", telemetry.ShortHash(addr),
						))
						s.Write([]byte("Too many sessions from your IP. Please try again later.\n"))
						s.Exit(1)
						return
					}
					defer sessionCounter.Release(addr)
					next(s)
				}
			},
			// Custom logging middleware (replaces wish/logging)
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					// Extract full session info for detailed logging
					info := telemetry.ExtractSessionInfo(s)
					logger.Debug("SSH session started", telemetry.Ctx(
						"session_hash", info.SessionHash,
						"user_hash", info.UserHash,
						"ip_hash", info.IPHash,
						"terminal", info.Terminal,
						"key_type", info.PublicKeyType,
						"term_program", info.EnvTermProgram,
						"shell", info.EnvShell,
					))
					next(s)
				}
			},
		),
	)
	if err != nil {
		logger.Error("Failed to create server", telemetry.Ctx("error", err.Error()))
		os.Exit(1)
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("SSH server starting...")
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Error("Server error", telemetry.Ctx("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("SSH server ready", telemetry.Ctx("command", "ssh -p "+port+" localhost"))
	<-done

	logger.Info("Shutting down...")
	analytics.TrackServerStop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error", telemetry.Ctx("error", err.Error()))
	}

	logger.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getContentPath() string {
	paths := []string{
		"../../packages/shared-content",
		"../packages/shared-content",
		"./packages/shared-content",
	}

	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(abs, "resume.json")); err == nil {
			return abs
		}
	}

	return "./packages/shared-content"
}

// SessionCounter tracks sessions per IP for rate limiting
type SessionCounter struct {
	counts   map[string]int
	maxPerIP int
}

func NewSessionCounter(maxPerIP int) *SessionCounter {
	return &SessionCounter{
		counts:   make(map[string]int),
		maxPerIP: maxPerIP,
	}
}

func (sc *SessionCounter) Acquire(addr string) bool {
	ip := addr
	if colonIdx := len(addr) - 1; colonIdx > 0 {
		for i := len(addr) - 1; i >= 0; i-- {
			if addr[i] == ':' {
				ip = addr[:i]
				break
			}
		}
	}

	if sc.counts[ip] >= sc.maxPerIP {
		return false
	}
	sc.counts[ip]++
	return true
}

func (sc *SessionCounter) Release(addr string) {
	ip := addr
	if colonIdx := len(addr) - 1; colonIdx > 0 {
		for i := len(addr) - 1; i >= 0; i-- {
			if addr[i] == ':' {
				ip = addr[:i]
				break
			}
		}
	}

	if sc.counts[ip] > 0 {
		sc.counts[ip]--
	}
	if sc.counts[ip] == 0 {
		delete(sc.counts, ip)
	}
}
