package dummy_test

import (
	"context"
	"testing"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
	"github.com/coditary/wuji/internal/driver/dummy"
)

func TestDummyDriverGenerateText(t *testing.T) {
	d := dummy.New()

	if !d.HasCapability(capability.TextGeneration) {
		t.Fatal("dummy driver should support text generation")
	}

	resp, err := d.GenerateText(context.Background(), driver.TextRequest{
		Prompt:      "hello",
		MaxTokens:   64,
		Temperature: 0.7,
	})
	if err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	if resp.Text == "" {
		t.Fatal("expected non-empty response")
	}
}

func TestDummyDriverInfo(t *testing.T) {
	d := dummy.New()
	info := d.Info()

	if info.ID != dummy.DriverID {
		t.Fatalf("expected ID %q, got %q", dummy.DriverID, info.ID)
	}
}
