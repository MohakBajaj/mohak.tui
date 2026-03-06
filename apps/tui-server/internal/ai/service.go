package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/telemetry"
)

const maxMessageLength = 2000

// Analytics captures AI-specific telemetry without coupling to a concrete implementation.
type Analytics interface {
	TrackAIRequest(sessionID string, messageLength int, historyLength int, model string)
	TrackAIResponse(sessionID string, durationMs int64, model string, success bool)
	TrackAIError(sessionID string, errorMsg string, errorType string)
	TrackAIRateLimit(sessionID string, remaining int)
}

// Config configures the in-process AI chat service.
type Config struct {
	Provider         Provider
	Logger           *telemetry.Logger
	Analytics        Analytics
	PromptBuilder    *PromptBuilder
	Model            string
	MaxTokens        int
	Temperature      float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
	MaxHistoryLength int
	RateLimitMax     int
	RateLimitWindow  time.Duration
}

// Service orchestrates validation, prompting, rate limiting, and provider calls.
type Service struct {
	provider  Provider
	logger    *telemetry.Logger
	analytics Analytics
	prompts   *PromptBuilder

	model            string
	maxTokens        int
	temperature      float64
	topP             float64
	frequencyPenalty float64
	presencePenalty  float64
	maxHistoryLength int
	rateLimitMax     int
	rateLimitWindow  time.Duration

	mu        sync.Mutex
	rateLimit map[string]rateLimitEntry
}

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

// NewService creates a new in-process AI service.
func NewService(cfg Config) *Service {
	return &Service{
		provider:         cfg.Provider,
		logger:           cfg.Logger,
		analytics:        cfg.Analytics,
		prompts:          cfg.PromptBuilder,
		model:            cfg.Model,
		maxTokens:        cfg.MaxTokens,
		temperature:      cfg.Temperature,
		topP:             cfg.TopP,
		frequencyPenalty: cfg.FrequencyPenalty,
		presencePenalty:  cfg.PresencePenalty,
		maxHistoryLength: cfg.MaxHistoryLength,
		rateLimitMax:     cfg.RateLimitMax,
		rateLimitWindow:  cfg.RateLimitWindow,
		rateLimit:        make(map[string]rateLimitEntry),
	}
}

// ChatStream validates, rate limits, builds the prompt, and streams the provider response.
func (s *Service) ChatStream(
	ctx context.Context,
	sessionID string,
	message string,
	history []Message,
	callback StreamCallback,
) error {
	requestStart := time.Now()

	if message == "" {
		return errors.New("message is required")
	}
	if len(message) > maxMessageLength {
		return fmt.Errorf("message too long (max %d characters)", maxMessageLength)
	}

	processedMessage := PreprocessMessage(message)
	intent := DetectQueryIntent(processedMessage)
	trimmedHistory := trimHistory(history, s.maxHistoryLength)

	if s.analytics != nil {
		s.analytics.TrackAIRequest(sessionID, len(processedMessage), len(trimmedHistory), s.model)
	}

	s.logger.Info("AI request received", telemetry.Ctx(
		"session_hash", sessionID,
		"message_length", len(processedMessage),
		"history_length", len(trimmedHistory),
		"intent", string(intent),
		"model", s.model,
	))

	remaining, allowed := s.checkRateLimit(sessionID)
	if !allowed {
		s.logger.Warn("AI rate limit exceeded", telemetry.Ctx("session_hash", sessionID))
		if s.analytics != nil {
			s.analytics.TrackAIRateLimit(sessionID, 0)
			s.analytics.TrackAIError(sessionID, "rate limit exceeded", "rate_limit")
		}
		return errors.New("rate limit exceeded - please wait before sending more messages")
	}

	messages := make([]CompletionMessage, 0, len(trimmedHistory)+2)
	messages = append(messages, CompletionMessage{
		Role:    "system",
		Content: s.prompts.BuildSystemPrompt(processedMessage),
	})
	for _, historyMessage := range trimmedHistory {
		messages = append(messages, CompletionMessage{
			Role:    historyMessage.Role,
			Content: historyMessage.Content,
		})
	}
	messages = append(messages, CompletionMessage{
		Role:    "user",
		Content: processedMessage,
	})

	err := s.provider.StreamChat(ctx, CompletionRequest{
		SessionID:        sessionID,
		Model:            s.model,
		Messages:         messages,
		MaxTokens:        s.maxTokens,
		Temperature:      s.temperature,
		TopP:             s.topP,
		FrequencyPenalty: s.frequencyPenalty,
		PresencePenalty:  s.presencePenalty,
	}, callback)
	if err != nil {
		errorType := "provider_error"
		if errors.Is(err, context.Canceled) {
			errorType = "cancelled"
		}
		s.logger.Error("AI response failed", telemetry.Ctx(
			"session_hash", sessionID,
			"error", err.Error(),
			"error_type", errorType,
			"rate_limit_remaining", remaining,
		))
		if s.analytics != nil {
			s.analytics.TrackAIResponse(sessionID, time.Since(requestStart).Milliseconds(), s.model, false)
			s.analytics.TrackAIError(sessionID, err.Error(), errorType)
		}
		return err
	}

	if s.analytics != nil {
		s.analytics.TrackAIResponse(sessionID, time.Since(requestStart).Milliseconds(), s.model, true)
	}

	s.logger.Info("AI response completed", telemetry.Ctx(
		"session_hash", sessionID,
		"duration_ms", time.Since(requestStart).Milliseconds(),
		"rate_limit_remaining", remaining,
		"intent", string(intent),
		"model", s.model,
	))

	return nil
}

func (s *Service) checkRateLimit(sessionID string) (remaining int, allowed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.rateLimit {
		if now.After(entry.resetAt) {
			delete(s.rateLimit, key)
		}
	}

	entry, ok := s.rateLimit[sessionID]
	if !ok || now.After(entry.resetAt) {
		s.rateLimit[sessionID] = rateLimitEntry{
			count:   1,
			resetAt: now.Add(s.rateLimitWindow),
		}
		return s.rateLimitMax - 1, true
	}

	if entry.count >= s.rateLimitMax {
		return 0, false
	}

	entry.count++
	s.rateLimit[sessionID] = entry
	return s.rateLimitMax - entry.count, true
}

func trimHistory(history []Message, maxHistoryLength int) []Message {
	if maxHistoryLength <= 0 || len(history) <= maxHistoryLength {
		return history
	}

	return history[len(history)-maxHistoryLength:]
}
