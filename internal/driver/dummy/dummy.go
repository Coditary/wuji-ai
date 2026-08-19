package dummy

import (
	"context"
	"fmt"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

const (
	DriverID   = "dummy"
	DriverName = "Dummy Driver"
)

// Driver is a placeholder backend for development and testing.
type Driver struct{}

func New() *Driver {
	return &Driver{}
}

func (d *Driver) Info() driver.Info {
	return driver.Info{
		ID:          DriverID,
		Name:        DriverName,
		Version:     "0.1.0",
		Description: "A dummy driver that returns placeholder text for walking skeleton testing.",
		Capabilities: []capability.Type{
			capability.TextGeneration,
		},
		Remote: false,
	}
}

func (d *Driver) Capabilities() []capability.Type {
	return d.Info().Capabilities
}

func (d *Driver) HasCapability(c capability.Type) bool {
	for _, cap := range d.Capabilities() {
		if cap == c {
			return true
		}
	}
	return false
}

func (d *Driver) GenerateText(_ context.Context, req driver.TextRequest) (*driver.TextResponse, error) {
	response := fmt.Sprintf(
		"[dummy] Generated response for prompt: %q (max_tokens=%d, temperature=%.1f)",
		req.Prompt,
		req.MaxTokens,
		req.Temperature,
	)

	return &driver.TextResponse{
		Text:         response,
		TokensUsed:   len(response) / 4,
		FinishReason: "stop",
	}, nil
}

func (d *Driver) Close() error {
	return nil
}
