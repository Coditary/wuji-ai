package cli

import (
	"testing"

	"github.com/spf13/cobra"

	wujicfg "github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/core"
	"github.com/coditary/wuji-core/pkg/driver/dummy"
)

func TestResolveImageLoRAsNamedAndDefault(t *testing.T) {
	cfg := wujicfg.Config{
		Root: t.TempDir(),
		LoRAs: wujicfg.LoRAConfig{
			Default: []string{"detail"},
			Library: map[string]wujicfg.LoRALibraryEntry{
				"style":  {Path: "anime_style.safetensors"},
				"detail": {Path: "detail.safetensors", Weight: 0.5},
			},
		},
	}
	app := &App{Core: mustTestCore(t, cfg), Config: &cfg}

	cmd := &cobra.Command{}
	var loras []string
	styleWeight := ""
	state := loraFlagState{namedWeights: map[string]*string{"style": &styleWeight}}
	cmd.Flags().StringSliceVar(&loras, "lora", nil, "")
	cmd.Flags().Lookup("lora").NoOptDefVal = wujicfg.LoRAUseDefaultsSentinel()
	cmd.Flags().StringVar(&styleWeight, "lora-style", "", "")

	if err := cmd.ParseFlags([]string{"--lora", "--lora-style", "70%"}); err != nil {
		t.Fatal(err)
	}

	refs, err := resolveLoRAs(cmd, app, loras, 0, state)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 {
		t.Fatalf("got %d refs, want 2", len(refs))
	}
	if refs[0].Path != "detail.safetensors" || refs[0].Weight != 0.5 {
		t.Fatalf("default ref = %+v", refs[0])
	}
	if refs[1].Path != "anime_style.safetensors" || refs[1].Weight != 0.7 {
		t.Fatalf("style ref = %+v", refs[1])
	}
}

func mustTestCore(t *testing.T, cfg wujicfg.Config) *core.Core {
	t.Helper()
	if cfg.DefaultDriver == "" {
		cfg.DefaultDriver = dummy.DriverID
	}
	c, err := core.New(core.Config{AppConfig: &cfg, DefaultDriverID: cfg.DefaultDriver, Lazy: true})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}
	return c
}
