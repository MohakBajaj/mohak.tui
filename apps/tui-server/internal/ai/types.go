package ai

import "context"

// Message represents a chat message exchanged with the model.
type Message struct {
	Role    string
	Content string
}

// StreamCallback is called for each streamed content chunk.
type StreamCallback func(chunk string) error

// ChatService is the interface consumed by the Bubble Tea model.
type ChatService interface {
	ChatStream(ctx context.Context, sessionID, message string, history []Message, callback StreamCallback) error
}

// Provider is a model backend that can stream a response.
type Provider interface {
	StreamChat(ctx context.Context, request CompletionRequest, callback StreamCallback) error
}

// CompletionMessage is the upstream message format sent to providers.
type CompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest contains the provider-agnostic generation request.
type CompletionRequest struct {
	SessionID        string
	Model            string
	Messages         []CompletionMessage
	MaxTokens        int
	Temperature      float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
}

// QueryIntent classifies the user message for prompt selection.
type QueryIntent string

const (
	IntentGreeting     QueryIntent = "greeting"
	IntentAbout        QueryIntent = "about"
	IntentExperience   QueryIntent = "experience"
	IntentSkills       QueryIntent = "skills"
	IntentProjects     QueryIntent = "projects"
	IntentContact      QueryIntent = "contact"
	IntentEducation    QueryIntent = "education"
	IntentAchievements QueryIntent = "achievements"
	IntentGeneral      QueryIntent = "general"
	IntentMeta         QueryIntent = "meta"
)
