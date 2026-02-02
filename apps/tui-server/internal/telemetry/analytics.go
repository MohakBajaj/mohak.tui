package telemetry

import (
	"os"
	"sync"
	"time"

	"github.com/posthog/posthog-go"
)

// Analytics provides PostHog integration for the TUI server
type Analytics struct {
	client posthog.Client
	logger *Logger
	mu     sync.Mutex
}

// Event types
const (
	EventSessionConnected    = "tui_session_connected"
	EventSessionDisconnected = "tui_session_disconnected"
	EventViewChanged         = "tui_view_changed"
	EventCommandExecuted     = "tui_command_executed"
	EventChatSent            = "tui_chat_sent"
	EventChatReceived        = "tui_chat_received"
	EventChatError           = "tui_chat_error"
	EventServerStart         = "tui_server_start"
	EventServerStop          = "tui_server_stop"
)

// NewAnalytics creates a new Analytics instance
func NewAnalytics(logger *Logger) *Analytics {
	apiKey := os.Getenv("POSTHOG_API_KEY")
	host := os.Getenv("POSTHOG_HOST")
	if host == "" {
		host = "https://us.i.posthog.com"
	}

	a := &Analytics{
		logger: logger,
	}

	if apiKey == "" {
		logger.Warn("PostHog API key not set, analytics disabled")
		return a
	}

	client, err := posthog.NewWithConfig(apiKey, posthog.Config{
		Endpoint:  host,
		BatchSize: 10,
		Interval:  5 * time.Second,
	})

	if err != nil {
		logger.Error("Failed to initialize PostHog", Ctx("error", err.Error()))
		return a
	}

	a.client = client
	logger.Info("PostHog analytics initialized", Ctx("host", host))

	return a
}

// capture sends an event to PostHog
func (a *Analytics) capture(event string, distinctID string, properties posthog.Properties) {
	if a.client == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if properties == nil {
		properties = posthog.NewProperties()
	}

	properties.Set("service", "tui-server")
	properties.Set("environment", getEnv("NODE_ENV", "development"))

	err := a.client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: properties,
	})

	if err != nil {
		a.logger.Error("Failed to capture analytics event", Ctx(
			"event", event,
			"error", err.Error(),
		))
	}
}

// TrackSessionConnected tracks when a user connects via SSH
func (a *Analytics) TrackSessionConnected(sessionID string, props map[string]interface{}) {
	properties := posthog.NewProperties()
	for k, v := range props {
		properties.Set(k, v)
	}
	a.capture(EventSessionConnected, sessionID, properties)
}

// TrackSessionDisconnected tracks when a user disconnects
func (a *Analytics) TrackSessionDisconnected(sessionID string, durationMs int64) {
	a.capture(EventSessionDisconnected, sessionID, posthog.NewProperties().
		Set("duration_ms", durationMs))
}

// TrackViewChanged tracks navigation between views
func (a *Analytics) TrackViewChanged(sessionID string, fromView, toView string) {
	a.capture(EventViewChanged, sessionID, posthog.NewProperties().
		Set("from_view", fromView).
		Set("to_view", toView))
}

// TrackCommandExecuted tracks slash commands
func (a *Analytics) TrackCommandExecuted(sessionID string, command string) {
	a.capture(EventCommandExecuted, sessionID, posthog.NewProperties().
		Set("command", command))
}

// TrackChatSent tracks when user sends a chat message
func (a *Analytics) TrackChatSent(sessionID string, messageLength int) {
	a.capture(EventChatSent, sessionID, posthog.NewProperties().
		Set("message_length", messageLength))
}

// TrackChatReceived tracks when AI response is received
func (a *Analytics) TrackChatReceived(sessionID string, responseLength int, durationMs int64) {
	a.capture(EventChatReceived, sessionID, posthog.NewProperties().
		Set("response_length", responseLength).
		Set("duration_ms", durationMs))
}

// TrackChatError tracks chat errors
func (a *Analytics) TrackChatError(sessionID string, errorMsg string) {
	a.capture(EventChatError, sessionID, posthog.NewProperties().
		Set("error", errorMsg))
}

// TrackServerStart tracks server startup
func (a *Analytics) TrackServerStart(host, port string) {
	a.capture(EventServerStart, "system", posthog.NewProperties().
		Set("host", host).
		Set("port", port))
}

// TrackServerStop tracks server shutdown
func (a *Analytics) TrackServerStop() {
	a.capture(EventServerStop, "system", nil)
}

// Identify associates user properties with a session
func (a *Analytics) Identify(sessionID string, properties map[string]interface{}) {
	if a.client == nil {
		return
	}

	props := posthog.NewProperties()
	for k, v := range properties {
		props.Set(k, v)
	}

	a.client.Enqueue(posthog.Identify{
		DistinctId: sessionID,
		Properties: props,
	})
}

// Close shuts down the analytics client
func (a *Analytics) Close() error {
	if a.client == nil {
		return nil
	}

	a.logger.Info("Shutting down PostHog client")
	return a.client.Close()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
