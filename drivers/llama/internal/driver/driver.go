package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/coditary/wuji/drivers/llama/internal/client"
	"github.com/coditary/wuji/drivers/llama/internal/config"
	"github.com/coditary/wuji/drivers/llama/internal/server"
	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

const (
	ID   = "llama"
	Name = "llama.cpp"
)

type Driver struct {
	cfg    *config.Config
	server *server.Server
	ollama *client.Ollama
}

func New(cfg *config.Config) (*Driver, error) {
	if !cfg.ServerAvailable() {
		return nil, fmt.Errorf("inference server binary not found at %s", cfg.ServerBin)
	}
	return &Driver{
		cfg:    cfg,
		server: server.New(cfg),
		ollama: client.NewOllama(cfg.OllamaAPI),
	}, nil
}

func (d *Driver) Info() driver.Info {
	return driver.Info{
		ID:          ID,
		Name:        Name,
		Version:     "b10502",
		Description: "Local text generation via llama.cpp or Ollama API.",
		Capabilities: []capability.Type{
			capability.TextGeneration,
		},
	}
}

func (d *Driver) Capabilities() []capability.Type {
	return d.Info().Capabilities
}

func (d *Driver) GenerateText(ctx context.Context, req driver.TextRequest) (*driver.TextResponse, error) {
	modelRef := req.Model
	if modelRef == "" {
		models, err := d.cfg.ListModels()
		if err != nil {
			return nil, err
		}
		if len(models) == 1 {
			modelRef = models[0]
		}
	}

	modelPath, err := d.cfg.ResolveModel(modelRef)
	if err != nil {
		return nil, err
	}

	temperature := req.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}

	// Use Ollama API when blob is not readable but mapped (common with system Ollama installs).
	if !d.cfg.ModelReadable(modelPath) {
		ollamaName, ok := d.cfg.ResolveOllamaName(modelRef)
		if !ok {
			return nil, fmt.Errorf("model %q not readable — run: sudo chmod -R g+rX /usr/share/ollama/.ollama", modelRef)
		}
		if !d.ollama.Available(ctx) {
			return nil, fmt.Errorf("ollama API not reachable at %s", d.cfg.OllamaAPI)
		}
		isThink := strings.Contains(strings.ToLower(ollamaName), "think")
		maxTokens := normalizeMaxTokens(req.MaxTokens, isThink)
		resp, err := d.ollama.Generate(ctx, ollamaName, req.Prompt, maxTokens, temperature, d.cfg.OllamaThink)
		if err != nil {
			return nil, err
		}
		text, err := resp.Answer()
		if err != nil {
			return nil, err
		}
		finishReason := "stop"
		if resp.DoneReason != "" {
			finishReason = resp.DoneReason
		}
		return &driver.TextResponse{
			Text: text, TokensUsed: resp.EvalCount, FinishReason: finishReason,
		}, nil
	}

	if err := d.server.EnsureRunning(ctx, modelPath); err != nil {
		return nil, err
	}

	maxTokens := normalizeMaxTokens(req.MaxTokens, false)

	c := d.server.Client()
	if c == nil {
		return nil, fmt.Errorf("inference client not ready")
	}

	resp, err := c.Complete(ctx, req.Prompt, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	finishReason := "stop"
	if resp.StoppedLimit {
		finishReason = "length"
	}

	return &driver.TextResponse{
		Text: resp.Content, TokensUsed: resp.TokensPredicted, FinishReason: finishReason,
	}, nil
}

func (d *Driver) Close() error {
	return d.server.Stop()
}

func (d *Driver) ListModels() ([]string, error) {
	return d.cfg.ListModels()
}

func (d *Driver) ModelsDir() string {
	return d.cfg.ModelsDir
}

// normalizeMaxTokens maps CLI values to backend limits.
// -1 = unlimited (Ollama/llama.cpp: generate until stop or context full).
// 0  = default (1024). Think models get at least 2048 unless unlimited.
func normalizeMaxTokens(requested int, isThink bool) int {
	if requested == -1 {
		return -1
	}
	if requested <= 0 {
		requested = 1024
	}
	if isThink && requested < 2048 {
		return 2048
	}
	return requested
}
