package cli

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), runErr
}

func TestListDrivers(t *testing.T) {
	app := newTestApp(t)
	out, err := captureStdout(func() error {
		return printDriverList(app)
	})
	if err != nil {
		t.Fatalf("printDriverList: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("dummy")) {
		t.Fatalf("expected dummy driver in output, got: %s", out)
	}
}

func TestListCapabilities(t *testing.T) {
	out, err := captureStdout(func() error {
		printCapabilityList()
		return nil
	})
	if err != nil {
		t.Fatalf("printCapabilityList: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("text")) {
		t.Fatalf("expected text capability in output, got: %s", out)
	}
}

func TestInfoDriver(t *testing.T) {
	app := newTestApp(t)
	out, err := captureStdout(func() error {
		return printDriverInfo(app, "dummy")
	})
	if err != nil {
		t.Fatalf("printDriverInfo: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("dummy")) {
		t.Fatalf("expected dummy driver info, got: %s", out)
	}
}

func TestInfoConfig(t *testing.T) {
	app := newTestApp(t)
	out, err := captureStdout(func() error {
		return printEffectiveConfig(app.Config)
	})
	if err != nil {
		t.Fatalf("printEffectiveConfig: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("VLLM")) {
		t.Fatalf("expected config sections in output, got: %s", out)
	}
}

func TestListDatasets(t *testing.T) {
	app := newTestApp(t)
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	out, err := captureStdout(func() error {
		return printDatasetList(app, cmd)
	})
	if err != nil {
		t.Fatalf("printDatasetList: %v", err)
	}
	if out == "" {
		t.Fatal("expected dataset list output")
	}
}

func TestListOverview(t *testing.T) {
	app := newTestApp(t)
	out, err := captureStdout(func() error {
		return printListOverview(app)
	})
	if err != nil {
		t.Fatalf("printListOverview: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("wuji list drivers")) {
		t.Fatalf("expected overview hints, got: %s", out)
	}
}
