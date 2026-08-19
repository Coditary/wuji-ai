package core_test

import (
	"context"
	"testing"

	"github.com/coditary/wuji/internal/core"
	"github.com/coditary/wuji/internal/driver"
	"github.com/coditary/wuji/internal/driver/dummy"
)

func TestCoreGenerateText(t *testing.T) {
	c, err := core.New(core.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	resp, err := c.GenerateText(context.Background(), "", driver.TextRequest{
		Prompt:      "test prompt",
		MaxTokens:   128,
		Temperature: 0.5,
	})
	if err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	if resp.Text == "" {
		t.Fatal("expected non-empty text response")
	}
}

func TestCoreListDrivers(t *testing.T) {
	c, err := core.New(core.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	drivers := c.ListDrivers()
	if len(drivers) != 1 {
		t.Fatalf("expected 1 driver, got %d", len(drivers))
	}

	if drivers[0].ID != dummy.DriverID {
		t.Fatalf("expected driver %q, got %q", dummy.DriverID, drivers[0].ID)
	}
}
