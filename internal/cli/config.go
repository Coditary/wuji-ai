package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	wujicfg "github.com/coditary/wuji-core/pkg/config"
)

func newConfigCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Wuji configuration (.wuji/config.yaml)",
	}

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set a driver option and save to .wuji/config.yaml",
	}

	setCmd.AddCommand(
		&cobra.Command{
			Use:   "vllm <key> <value>",
			Short: "Set a vLLM driver option",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetVLLMField(app.Config.Root, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set drivers.vllm.%s = %s\n", args[0], args[1])
				return nil
			},
		},
		&cobra.Command{
			Use:   "llama <key> <value>",
			Short: "Set a llama driver option",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetLlamaField(app.Config.Root, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set drivers.llama.%s = %s\n", args[0], args[1])
				return nil
			},
		},
		&cobra.Command{
			Use:   "a1111 <key> <value>",
			Short: "Set an A1111 image/video driver option",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetA1111Field(app.Config.Root, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set drivers.a1111.%s = %s\n", args[0], args[1])
				return nil
			},
		},
		&cobra.Command{
			Use:   "driver <capability> <driver-id>",
			Short: "Set the default driver for a capability (text, image, video, …)",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetCapabilityDriver(app.Config.Root, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set capability_drivers.%s = %s\n", args[0], args[1])
				return nil
			},
		},
		&cobra.Command{
			Use:   "default-driver <driver-id>",
			Short: "Set the default driver (e.g. llama, vllm, dummy)",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg := app.Config
				cfg.DefaultDriver = args[0]
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set default_driver = %s\n", args[0])
				return nil
			},
		},
		&cobra.Command{
			Use:   "lora <alias> <path>",
			Short: "Register a named LoRA in loras.library",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetLoRALibrary(app.Config.Root, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set loras.library.%s = %s\n", args[0], args[1])
				return nil
			},
		},
		&cobra.Command{
			Use:   "lora-default <alias[,alias...]>",
			Short: "Set loras.default (used by bare --lora)",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetLoRADefault(app.Config.Root, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set loras.default = %s\n", args[0])
				return nil
			},
		},
		&cobra.Command{
			Use:   "lora-weight <weight>",
			Short: "Set loras.default_weight (0.7, 1, or 70%%)",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetLoRADefaultWeight(app.Config.Root, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set loras.default_weight = %s\n", args[0])
				return nil
			},
		},
		&cobra.Command{
			Use:   "resources <key> <value>",
			Short: "Set a resource budget option (total_ram_mb, total_vram_mb, idle_shutdown_seconds, …)",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wujicfg.SetResourcesField(app.Config.Root, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Set drivers.resources.%s = %s\n", args[0], args[1])
				return nil
			},
		},
	)

	cmd.AddCommand(
		&cobra.Command{
			Use:        "list",
			Short:      "Show effective configuration",
			Deprecated: "use `wuji info config` instead",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printEffectiveConfig(app.Config)
			},
		},
		&cobra.Command{
			Use:   "show",
			Short: "Show effective configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printEffectiveConfig(app.Config)
			},
		},
		setCmd,
	)

	return cmd
}

func displayOrDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}
