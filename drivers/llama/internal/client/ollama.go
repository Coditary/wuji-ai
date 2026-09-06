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
	NumPredict       int      `json:"num_predict,omitempty"`
	Temperature      float32  `json:"temperature,omitempty"`
	TopP             float32  `json:"top_p,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	MinP             float32  `json:"min_p,omitempty"`
	FrequencyPenalty float32  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32  `json:"presence_penalty,omitempty"`
	RepeatPenalty    float32  `json:"repeat_penalty,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	NumCtx           int      `json:"num_ctx,omitempty"`
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

func buildOllamaOptions(p GenerationParams) ollamaOptions {
	opts := ollamaOptions{
		NumPredict:  p.MaxTokens,
		Temperature: p.Temperature,
	}
	if p.TopP > 0 {
		opts.TopP = p.TopP
	}
	if p.TopK > 0 {
		opts.TopK = p.TopK
	}
	if p.MinP > 0 {
		opts.MinP = p.MinP
	}
	if p.FrequencyPenalty != 0 {
		opts.FrequencyPenalty = p.FrequencyPenalty
	}
	if p.PresencePenalty != 0 {
		opts.PresencePenalty = p.PresencePenalty
	}
	if p.RepetitionPenalty > 0 && p.RepetitionPenalty != 1.0 {
		opts.RepeatPenalty = p.RepetitionPenalty
	}
	if len(p.StopSequences) > 0 {
		opts.Stop = append([]string(nil), p.StopSequences...)
	}
	if p.Seed != nil {
		opts.Seed = p.Seed
	}
	if p.ContextWindow > 0 {
		opts.NumCtx = p.ContextWindow
	}
	return opts
}

func (o *Ollama) Generate(ctx context.Context, model string, p GenerationParams, think bool) (*ollamaChatResponse, error) {
	messages := make([]ollamaMessage, 0, 2)
	if p.SystemPrompt != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: p.SystemPrompt})
	}
	messages = append(messages, ollamaMessage{Role: "user", Content: p.Prompt})

	body := ollamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Think:    think,
		Options:  buildOllamaOptions(p),
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
