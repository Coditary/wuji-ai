package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/capability"
)

func runLoadInference(app *App, cmd *cobra.Command, args []string) error {
	driverID := app.resolveDriver(cmd, capability.TextGeneration)
	if len(args) > 0 {
		driverID = args[0]
	}
	model, _ := cmd.Flags().GetString("model")
	if err := app.Core.LoadInference(context.Background(), driverID, model, capability.TextGeneration); err != nil {
		return err
	}
	if model == "" {
		model = "(default)"
	}
	fmt.Fprintf(os.Stdout, "Loaded %s model %s\n", driverID, model)
	return nil
}

func runUnloadInference(app *App, args []string) error {
	target := "all"
	if len(args) > 0 {
		target = args[0]
	}
	if err := app.Core.UnloadInferenceDriver(context.Background(), target); err != nil {
		return err
	}
	if target == "all" {
		fmt.Fprintln(os.Stdout, "Unloaded all inference servers (llama, vllm).")
	} else {
		fmt.Fprintf(os.Stdout, "Unloaded %s.\n", target)
	}
	return nil
}

func runKillInference(app *App) error {
	if err := app.Core.UnloadAllInference(context.Background()); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Inference servers stopped (llama, vllm).")
	return nil
}

func newLoadCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load [driver]",
		Short: "Preload a model into memory (start inference server)",
		Long: `Manually load a model before running generate commands.

Uses the same scheduler as wuji text/image (evicts other models if needed).
Driver defaults to the text capability driver from config, or --driver global flag.

Examples:
  wuji load vllm
  wuji load llama --model my-chat.gguf
  wuji load -d vllm --model meta-llama/Llama-3.2-1B`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadInference(app, cmd, args)
		},
	}
	cmd.Flags().String("model", "", "model name or path (default: from driver config)")
	return cmd
}

func newUnloadCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "unload [driver|all]",
		Short: "Unload model(s) from memory and stop inference servers",
		Long: `Manually free RAM/VRAM by stopping inference servers.

Examples:
  wuji unload vllm
  wuji unload llama
  wuji unload all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnloadInference(app, args)
		},
	}
}

func newKillCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "kill",
		Short: "Stop all inference servers (llama, vllm)",
		Long:  "Alias for `wuji unload all`. Frees RAM/VRAM by stopping all managed inference servers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKillInference(app)
		},
	}
}
