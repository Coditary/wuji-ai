package driver

import (
	"context"

	"github.com/coditary/wuji/internal/capability"
)

// Info describes a registered driver.
type Info struct {
	ID           string
	Name         string
	Version      string
	Description  string
	Capabilities []capability.Type
	Remote       bool
	Endpoint     string
}

// TextRequest is the input for text generation.
type TextRequest struct {
	Prompt      string
	MaxTokens   int
	Temperature float32
}

// TextResponse is the output of text generation.
type TextResponse struct {
	Text         string
	TokensUsed   int
	FinishReason string
}

// Driver is the interface every backend must implement.
type Driver interface {
	Info() Info
	Capabilities() []capability.Type
	HasCapability(c capability.Type) bool
	GenerateText(ctx context.Context, req TextRequest) (*TextResponse, error)
	Close() error
}
