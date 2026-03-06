package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/content"
	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/telemetry"
)

func TestDetectQueryIntent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		query    string
		expected QueryIntent
	}{
		{name: "greeting", query: "hello there", expected: IntentGreeting},
		{name: "meta", query: "how does this tui work", expected: IntentMeta},
		{name: "projects", query: "what projects has he built", expected: IntentProjects},
		{name: "general", query: "tell me something interesting", expected: IntentGeneral},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if actual := DetectQueryIntent(testCase.query); actual != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}

func TestPreprocessMessage(t *testing.T) {
	t.Parallel()

	if actual := PreprocessMessage("skills"); actual != "Tell me about Mohak's skills" {
		t.Fatalf("unexpected single word expansion: %q", actual)
	}

	if actual := PreprocessMessage("hi pls tell me ur stack thx"); actual != "hi please tell me your stack thanks" {
		t.Fatalf("unexpected shorthand normalization: %q", actual)
	}
}

func TestPromptBuilderBuildSystemPrompt(t *testing.T) {
	t.Parallel()

	loader := content.NewLoader("")
	resume, err := loader.LoadResume()
	if err != nil {
		t.Fatalf("load resume: %v", err)
	}
	projects, err := loader.LoadProjects()
	if err != nil {
		t.Fatalf("load projects: %v", err)
	}
	bio, err := loader.LoadBio()
	if err != nil {
		t.Fatalf("load bio: %v", err)
	}

	builder := NewPromptBuilder(resume, projects, bio)
	prompt := builder.BuildSystemPrompt("how does this tui work")

	for _, expected := range []string{"## CONTEXT", "SSH TUI Portfolio", "Tech Stack:"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing %q", expected)
		}
	}
}

func TestServiceRateLimit(t *testing.T) {
	t.Parallel()

	loader := content.NewLoader("")
	resume, _ := loader.LoadResume()
	projects, _ := loader.LoadProjects()
	bio, _ := loader.LoadBio()

	service := NewService(Config{
		Provider:         stubProvider{},
		Logger:           telemetry.NewLogger("test"),
		PromptBuilder:    NewPromptBuilder(resume, projects, bio),
		Model:            "test-model",
		MaxTokens:        10,
		Temperature:      0.7,
		TopP:             0.9,
		FrequencyPenalty: 0.3,
		PresencePenalty:  0.1,
		MaxHistoryLength: 10,
		RateLimitMax:     1,
		RateLimitWindow:  time.Minute,
	})

	err := service.ChatStream(context.Background(), "session", "hello", nil, nil)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	err = service.ChatStream(context.Background(), "session", "hello again", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Fatalf("expected rate limit error, got %v", err)
	}
}

type stubProvider struct{}

func (stubProvider) StreamChat(_ context.Context, _ CompletionRequest, callback StreamCallback) error {
	if callback != nil {
		return callback("ok")
	}
	return nil
}
