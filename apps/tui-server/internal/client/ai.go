package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request payload for the AI gateway
type ChatRequest struct {
	Message   string    `json:"message"`
	SessionID string    `json:"sessionId"`
	History   []Message `json:"history,omitempty"`
}

// AIClient handles communication with the AI gateway
type AIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAIClient creates a new AI client
func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// StreamCallback is called for each chunk of the streaming response
type StreamCallback func(chunk string) error

// ChatStream sends a message and streams the response chunk by chunk
func (c *AIClient) ChatStream(ctx context.Context, sessionID, message string, history []Message, callback StreamCallback) error {
	req := ChatRequest{
		Message:   message,
		SessionID: sessionID,
		History:   history,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return errors.New("rate limit exceeded - please wait before sending more messages")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI gateway error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read the streaming response byte by byte for real-time streaming
	reader := bufio.NewReader(resp.Body)
	buf := make([]byte, 1)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		if n > 0 {
			if callback != nil {
				if err := callback(string(buf[:n])); err != nil {
					return err
				}
			}
		}
	}
}

// Chat sends a message and returns the full response (non-streaming)
func (c *AIClient) Chat(ctx context.Context, sessionID, message string, history []Message, callback StreamCallback) error {
	return c.ChatStream(ctx, sessionID, message, history, callback)
}

// Health checks if the AI gateway is available
func (c *AIClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
