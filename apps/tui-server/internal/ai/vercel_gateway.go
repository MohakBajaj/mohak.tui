package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/network"
)

const vercelGatewayBaseURL = "https://ai-gateway.vercel.sh/v1"

// VercelGatewayProvider streams chat completions from the Vercel AI Gateway.
type VercelGatewayProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewVercelGatewayProvider creates a Vercel AI Gateway provider.
func NewVercelGatewayProvider(apiKey string) *VercelGatewayProvider {
	return &VercelGatewayProvider{
		apiKey:  apiKey,
		baseURL: vercelGatewayBaseURL,
		httpClient: &http.Client{
			Timeout:   120 * time.Second,
			Transport: network.NewHTTPTransport(),
		},
	}
}

// StreamChat sends a streaming chat completion request and emits content deltas.
func (p *VercelGatewayProvider) StreamChat(
	ctx context.Context,
	request CompletionRequest,
	callback StreamCallback,
) error {
	if strings.TrimSpace(p.apiKey) == "" {
		return errors.New("AI_GATEWAY_API_KEY is required")
	}

	body, err := json.Marshal(openAIChatRequest{
		Model:            request.Model,
		Messages:         request.Messages,
		Stream:           true,
		MaxTokens:        request.MaxTokens,
		Temperature:      request.Temperature,
		TopP:             request.TopP,
		FrequencyPenalty: request.FrequencyPenalty,
		PresencePenalty:  request.PresencePenalty,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal provider request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create provider request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := p.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("failed to send provider request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		return errors.New("rate limit exceeded - please wait before sending more messages")
	}
	if response.StatusCode != http.StatusOK {
		return readProviderError(response)
	}

	return streamOpenAIChunks(ctx, response.Body, callback)
}

type openAIChatRequest struct {
	Model            string              `json:"model"`
	Messages         []CompletionMessage `json:"messages"`
	Stream           bool                `json:"stream"`
	MaxTokens        int                 `json:"max_tokens,omitempty"`
	Temperature      float64             `json:"temperature,omitempty"`
	TopP             float64             `json:"top_p,omitempty"`
	FrequencyPenalty float64             `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64             `json:"presence_penalty,omitempty"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type providerErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func readProviderError(response *http.Response) error {
	body, _ := io.ReadAll(response.Body)
	parsed := providerErrorResponse{}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		return fmt.Errorf("AI provider error (status %d): %s", response.StatusCode, parsed.Error.Message)
	}

	return fmt.Errorf("AI provider error (status %d): %s", response.StatusCode, strings.TrimSpace(string(body)))
}

func streamOpenAIChunks(ctx context.Context, body io.Reader, callback StreamCallback) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			return nil
		}

		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("failed to parse provider stream: %w", err)
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content == "" {
				continue
			}
			if callback != nil {
				if err := callback(choice.Delta.Content); err != nil {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading provider stream: %w", err)
	}

	return nil
}
