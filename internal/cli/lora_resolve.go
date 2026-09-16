package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	wujicfg "github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/driver"
)

type loraFlagState struct {
	namedWeights map[string]*string
}

func registerLoRAFlags(cmd *cobra.Command, app *App, state *loraFlagState) {
	state.namedWeights = make(map[string]*string)
	loraCfg := app.Config.ResolvedLoRAs()
	names := make([]string, 0, len(loraCfg.Library))
	for name := range loraCfg.Library {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		state.namedWeights[name] = new(string)
		flagName := "lora-" + name
		cmd.Flags().StringVar(state.namedWeights[name], flagName, "", fmt.Sprintf("LoRA %q from config (weight: 0.7 or 70%%)", name))
	}

	if flag := cmd.Flags().Lookup("lora"); flag != nil {
		flag.NoOptDefVal = wujicfg.LoRAUseDefaultsSentinel()
	}
}

func resolveLoRAs(cmd *cobra.Command, app *App, entries []string, fallbackWeight float32, state loraFlagState) ([]driver.LoRARef, error) {
	cfg := app.Config.ResolvedLoRAs()
	if fallbackWeight <= 0 {
		fallbackWeight = cfg.DefaultWeight
	}
	if fallbackWeight <= 0 {
		fallbackWeight = 1
	}

	var out []driver.LoRARef
	seen := make(map[string]struct{})

	appendRef := func(ref driver.LoRARef) error {
		key := ref.Path + fmt.Sprintf("@%.4f", ref.LoRAWeight())
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		out = append(out, ref)
		return nil
	}

	if cmd.Flags().Changed("lora") {
		for _, entry := range entries {
			if entry == wujicfg.LoRAUseDefaultsSentinel() {
				refs, err := cfg.ResolveDefault(fallbackWeight)
				if err != nil {
					return nil, err
				}
				for _, ref := range refs {
					if err := appendRef(ref); err != nil {
						return nil, err
					}
				}
				continue
			}
			ref, err := cfg.ResolveEntry(entry, fallbackWeight)
			if err != nil {
				return nil, err
			}
			if err := appendRef(ref); err != nil {
				return nil, err
			}
		}
	}

	for name, weightPtr := range state.namedWeights {
		flagName := "lora-" + name
		if !cmd.Flags().Changed(flagName) {
			continue
		}
		ref, err := cfg.ResolveLibraryName(name, *weightPtr, fallbackWeight)
		if err != nil {
			return nil, err
		}
		if err := appendRef(ref); err != nil {
			return nil, err
		}
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func describeLoRAConfig(cfg wujicfg.LoRAConfig) string {
	if len(cfg.Library) == 0 && len(cfg.Default) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("LORAS\n")
	if cfg.DefaultWeight > 0 && cfg.DefaultWeight != 1 {
		fmt.Fprintf(&b, "  default_weight\t%g\n", cfg.DefaultWeight)
	}
	if len(cfg.Default) > 0 {
		fmt.Fprintf(&b, "  default\t%s\n", strings.Join(cfg.Default, ", "))
	}
	names := make([]string, 0, len(cfg.Library))
	for name := range cfg.Library {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry := cfg.Library[name]
		if entry.Weight > 0 {
			fmt.Fprintf(&b, "  %s\t%s (weight %g)\n", name, entry.Path, entry.Weight)
		} else {
			fmt.Fprintf(&b, "  %s\t%s\n", name, entry.Path)
		}
	}
	return b.String()
}
