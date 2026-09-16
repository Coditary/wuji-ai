package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	wujicfg "github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/sysmem"
	"github.com/spf13/cobra"
)

const resourcesExplainText = `Automatic resource management
=============================

Wuji manages RAM and VRAM for every generate command automatically:
  wuji text, wuji image, wuji video, wuji audio

You do NOT need to enable anything manually. If no drivers.resources block
exists in .wuji/config.yaml, Wuji detects your system RAM (and GPU VRAM via
nvidia-smi) and uses conservative budgets (80%% RAM, 90%% VRAM).

What happens on each command
----------------------------

1. LOCK     A file lock (.wuji/runtime/scheduler.lock) coordinates parallel
            wuji processes so they don't fight over memory.

2. PLAN     Before the job runs, the scheduler checks whether the model fits
            into the RAM/VRAM budget. If not, it evicts other loaded models
            (e.g. unloads your chat model to make room for image generation).

3. RUN      Your command executes (text, image, …).

4. CLEANUP  Transient jobs (image, video, audio, mesh) unload their model
            after completion so memory is freed immediately.

5. RESTORE  With lazy_restore (default: true), evicted models are NOT
            reloaded immediately. They reload only when you run the matching
            command again (e.g. wuji text after wuji image).

Example: chat + image without OOM
-----------------------------------

  Terminal A: wuji text -p "hello"     → loads vllm chat model
  Terminal B: wuji image -p "a cat"  → evicts chat, runs image, unloads image
  Terminal A: wuji text -p "again"   → reloads chat on demand

State is persisted in .wuji/runtime/scheduler.json across CLI invocations.

Useful commands
---------------

  wuji resources status   budgets, loaded models, restore queue
  wuji load               preload a model (wuji load vllm --model …)
  wuji unload             free memory (wuji unload vllm | all)
  wuji kill               stop all inference servers
  wuji resources detect   show detected RAM/VRAM
  wuji resources init     write budgets to .wuji/config.yaml
  wuji config set resources total_ram_mb 51200

Daemon idle shutdown
--------------------

When running wuji serve, remote drivers (vllm, a1111, …) stay loaded after use.
After idle_shutdown_seconds with NO active wuji command (default: 1800 = 30 min), the
daemon stops the driver process and frees its socket.

Important: long-running work is protected. While wuji text/image/train/… is running
(even for hours), the driver is marked busy and will NOT be stopped. Shutdown only
happens after the command finishes and the idle timer expires.

Built-ins (dummy, local, ffmpeg) are never stopped. Set idle_shutdown_seconds: 0 to disable.

  wuji config set resources idle_shutdown_seconds 600

Tuning
------

  model_ram_mb / model_vram_mb  — override per-model memory estimates
  lazy_restore: false           — reload evicted models immediately (uses more RAM)
  evict_and_restore: false      — fail instead of evicting when budget exceeded
  idle_shutdown_seconds: 0      — keep remote drivers loaded until manual unload
`

func newResourcesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Automatic RAM/VRAM scheduling for all generate commands",
		Long: `Manage and inspect automatic memory scheduling.

Every wuji text/image/video/… command runs through the resource scheduler
when a RAM or VRAM budget is available (auto-detected or configured).

Run 'wuji resources explain' for a full walkthrough.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "explain",
			Short: "How automatic resource management works",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Fprint(os.Stdout, resourcesExplainText)
				return nil
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show resource budget and loaded models",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg := app.Core.ResourcesConfig()
				state, err := app.Core.ResourceStatus()
				if err != nil {
					return err
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				if cfg.Enabled() {
					fmt.Fprintln(w, "AUTO_MANAGED\ttrue (every wuji text/image/video/… command)")
					fmt.Fprintf(w, "BUDGET_SOURCE\t%s\n", cfg.BudgetSource())
					if cfg.TotalVRAMMB > 0 {
						source := "config"
						if cfg.VRAMBudgetFromAuto {
							source = "auto-detected"
						}
						fmt.Fprintf(w, "VRAM_BUDGET\t%d MB (%s)\n", cfg.TotalVRAMMB, source)
					}
					if cfg.TotalRAMMB > 0 {
						source := "config"
						if cfg.RAMBudgetFromAuto {
							source = "auto-detected"
						}
						fmt.Fprintf(w, "RAM_BUDGET\t%d MB (%s)\n", cfg.TotalRAMMB, source)
					}
					fmt.Fprintf(w, "EVICT_AND_RESTORE\t%t\n", cfg.EvictAndRestore)
					fmt.Fprintf(w, "LAZY_RESTORE\t%t\n", cfg.LazyRestore)
					fmt.Fprintf(w, "UNLOAD_TRANSIENT\t%t\n", cfg.UnloadTransientModels)
					if cfg.IdleShutdownSeconds > 0 {
						fmt.Fprintf(w, "IDLE_SHUTDOWN\t%d seconds\n", cfg.IdleShutdownSeconds)
					} else {
						fmt.Fprintln(w, "IDLE_SHUTDOWN\tdisabled")
					}
				} else {
					fmt.Fprintln(w, "AUTO_MANAGED\tfalse")
					fmt.Fprintln(w, "STATUS\tdisabled — could not detect RAM or VRAM")
					printResourceHints(w)
				}
				fmt.Fprintln(w, "")

				if len(state.Loaded) == 0 {
					fmt.Fprintln(w, "LOADED\t(none)")
				} else {
					fmt.Fprintln(w, "LOADED")
					for _, slot := range state.Loaded {
						fmt.Fprintf(w, "  %s\t%s\t%s\t%d MB VRAM\t%d MB RAM\n",
							slot.DriverID, slot.Model, slot.Capability, slot.VRAMMB, slot.RAMMB)
					}
				}

				if len(state.RestoreStack) > 0 {
					fmt.Fprintln(w, "")
					fmt.Fprintln(w, "RESTORE_QUEUE (reload on next matching command)")
					for _, snap := range state.RestoreStack {
						fmt.Fprintf(w, "  %s\t%s\t%s\n", snap.DriverID, snap.Model, snap.Capability)
					}
				}
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "HINT\trun `wuji resources explain` for details")
				return w.Flush()
			},
		},
		&cobra.Command{
			Use:   "detect",
			Short: "Show detected system RAM and GPU VRAM",
			RunE: func(cmd *cobra.Command, args []string) error {
				totalRAM := sysmem.TotalRAMMB()
				totalVRAM := sysmem.TotalVRAMMB()
				suggestRAM, suggestVRAM := sysmem.SuggestedBudget()

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				if totalRAM > 0 {
					fmt.Fprintf(w, "SYSTEM_RAM\t%d MB\n", totalRAM)
					fmt.Fprintf(w, "SUGGESTED_RAM_BUDGET\t%d MB (80%%)\n", suggestRAM)
				} else {
					fmt.Fprintln(w, "SYSTEM_RAM\tunknown")
				}
				if totalVRAM > 0 {
					fmt.Fprintf(w, "GPU_VRAM\t%d MB\n", totalVRAM)
					fmt.Fprintf(w, "SUGGESTED_VRAM_BUDGET\t%d MB (90%%)\n", suggestVRAM)
				} else {
					fmt.Fprintln(w, "GPU_VRAM\tnot detected (no nvidia-smi?)")
				}
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "AUTO_MANAGED\ttrue when RAM or VRAM is detectable (no config required)")
				fmt.Fprintln(w, "HINT\trun `wuji resources init` to persist budgets to .wuji/config.yaml")
				return w.Flush()
			},
		},
		&cobra.Command{
			Use:   "init",
			Short: "Detect RAM/VRAM and write budgets to .wuji/config.yaml",
			RunE: func(cmd *cobra.Command, args []string) error {
				res, err := wujicfg.InitResources(app.Config.Root)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Resource budgets saved to .wuji/config.yaml\n")
				fmt.Fprintln(os.Stdout, "Scheduling stays automatic — this only pins the detected budgets.")
				if res.TotalRAMMB > 0 {
					fmt.Fprintf(os.Stdout, "  total_ram_mb: %d\n", res.TotalRAMMB)
				}
				if res.TotalVRAMMB > 0 {
					fmt.Fprintf(os.Stdout, "  total_vram_mb: %d\n", res.TotalVRAMMB)
				}
				return nil
			},
		},
	)

	loadCmd := newLoadCmd(app)
	loadCmd.Deprecated = "use `wuji load` instead"

	unloadCmd := newUnloadCmd(app)
	unloadCmd.Deprecated = "use `wuji unload` instead"

	killCmd := newKillCmd(app)
	killCmd.Deprecated = "use `wuji kill` instead"

	cmd.AddCommand(loadCmd, unloadCmd, killCmd)

	return cmd
}

func printResourceHints(w *tabwriter.Writer) {
	ram := sysmem.TotalRAMMB()
	if ram > 0 {
		suggested := ram * 80 / 100
		fmt.Fprintf(w, "SYSTEM_RAM\t%d MB (suggested budget: %d MB)\n", ram, suggested)
		fmt.Fprintf(w, "HINT\trun `wuji resources init` or set drivers.resources.total_ram_mb: %d\n", suggested)
	} else {
		fmt.Fprintln(w, "HINT\tset drivers.resources.total_ram_mb in .wuji/config.yaml")
	}
}
