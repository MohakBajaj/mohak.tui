package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/app"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/client"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/content"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/theme"
)

const (
	defaultHost      = "0.0.0.0"
	defaultPort      = "2222"
	idleTimeout      = 10 * time.Minute
	maxSessionsPerIP = 5
)

func main() {
	// Configuration from environment
	host := getEnv("SSH_HOST", defaultHost)
	port := getEnv("SSH_PORT", defaultPort)
	aiGatewayURL := getEnv("AI_GATEWAY_URL", "http://localhost:3001")
	contentPath := getEnv("CONTENT_PATH", getContentPath())

	log.Printf("Starting SSH server on %s:%s", host, port)
	log.Printf("AI Gateway URL: %s", aiGatewayURL)
	log.Printf("Content path: %s", contentPath)

	// Load content
	contentLoader := content.NewLoader(contentPath)

	resume, err := contentLoader.LoadResume()
	if err != nil {
		log.Fatalf("Failed to load resume: %v", err)
	}

	projects, err := contentLoader.LoadProjects()
	if err != nil {
		log.Fatalf("Failed to load projects: %v", err)
	}

	bio, err := contentLoader.LoadBio()
	if err != nil {
		log.Fatalf("Failed to load bio: %v", err)
	}

	// Create AI client
	aiClient := client.NewAIClient(aiGatewayURL)

	// Check AI gateway health (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := aiClient.Health(ctx); err != nil {
			log.Printf("Warning: AI gateway not available: %v", err)
		} else {
			log.Printf("AI gateway connected")
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
				// Get terminal size
				pty, _, active := s.Pty()
				if !active {
					log.Printf("No PTY for session %s", s.RemoteAddr())
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

				// Create session-specific theme manager
				themeManager := theme.NewManager(width, height)

				// Generate session ID for rate limiting
				sessionID := s.RemoteAddr().String()

				// Create model
				model := app.NewModel(app.Config{
					ThemeManager: themeManager,
					Resume:       resume,
					Projects:     projects,
					Bio:          bio,
					AIClient:     aiClient,
					SessionID:    sessionID,
					Width:        width,
					Height:       height,
				})

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
						log.Printf("Rate limited: %s", addr)
						s.Write([]byte("Too many sessions from your IP. Please try again later.\n"))
						s.Exit(1)
						return
					}
					defer sessionCounter.Release(addr)
					next(s)
				}
			},
			// Logging middleware
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("SSH server starting...")
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Printf("SSH server ready: ssh -p %s localhost", port)
	<-done

	log.Printf("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getContentPath() string {
	// Try to find the shared-content package relative to the executable
	// or use an environment variable
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

	// Default fallback
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
	// Extract IP from addr (format: "ip:port")
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
