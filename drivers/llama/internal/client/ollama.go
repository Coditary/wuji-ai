package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Think    bool            `json:"think"`
	Options  ollamaOptions   `json:"options"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	NumPredict  int     `json:"num_predict"`
	Temperature float32 `json:"temperature"`
}

type ollamaChatResponse struct {
	Message struct {
		Content  string `json:"content"`
		Thinking string `json:"thinking"`
	} `json:"message"`
	EvalCount  int    `json:"eval_count"`
	DoneReason string `json:"done_reason"`
}

func (r *ollamaChatResponse) Answer() (string, error) {
	if r.Message.Content != "" {
		return r.Message.Content, nil
	}
	if r.Message.Thinking != "" {
		if r.DoneReason == "length" {
			return "", fmt.Errorf("token limit reached while the model was still thinking — increase --max-tokens (e.g. 2048) or use a non-think model")
		}
		return "", fmt.Errorf("model produced no final answer (only internal reasoning)")
	}
	return "", fmt.Errorf("empty response from model")
}

// Ollama talks to a running Ollama instance via its HTTP API.
type Ollama struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllama(baseURL string) *Ollama {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	return &Ollama{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (o *Ollama) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL, nil)
	if err != nil {
		return false
	}
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (o *Ollama) Generate(ctx context.Context, model, prompt string, maxTokens int, temperature float32, think bool) (*ollamaChatResponse, error) {
	body := ollamaChatRequest{
		Model: model,
		Messages: []ollamaMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
		Think:  think,
		Options: ollamaOptions{
			NumPredict:  maxTokens,
			Temperature: temperature,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama chat failed (%d): %s", resp.StatusCode, string(raw))
	}

	var result ollamaChatResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode ollama response: %w", err)
	}
	return &result, nil
}
