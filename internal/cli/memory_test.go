package cli

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestKillInference(t *testing.T) {
	app := newTestApp(t)
	if err := runKillInference(app); err != nil {
		t.Fatalf("runKillInference: %v", err)
	}
}

func TestUnloadInferenceAll(t *testing.T) {
	app := newTestApp(t)
	if err := runUnloadInference(app, nil); err != nil {
		t.Fatalf("runUnloadInference: %v", err)
	}
}

func TestLoadInferenceDefaultDriver(t *testing.T) {
	app := newTestApp(t)
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	if err := runLoadInference(app, cmd, nil); err != nil {
		t.Fatalf("runLoadInference: %v", err)
	}
}
