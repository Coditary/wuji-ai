package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/core"
	"github.com/coditary/wuji-core/pkg/version"
)

// App holds the CLI application state.
type App struct {
	Core    core.Runtime
	Config  *config.Config
	Root    *cobra.Command
	cleanup func()
}

// New creates the root CLI application.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	app := &App{Config: cfg}

	if isLocalCommand(os.Args) {
		// catalog sync, provider login, etc. do not need the core daemon.
	} else {
		rt, cleanup, err := core.ConnectRuntime(context.Background(), cfg)
		if err != nil {
			return nil, fmt.Errorf("connect to core daemon: %w (run `wuji-core` or set WUJI_CORE_BIN)", err)
		}
		app.Core = rt
		app.cleanup = cleanup
	}

	root := &cobra.Command{
		Use:   "wuji",
		Short: "Unified CLI for AI backends and APIs",
		Long: `Wuji is a unified command-line interface for AI workloads.

Commands connect to the local Wuji core daemon (start it with: wuji-core).
Run capability commands directly: wuji text, wuji image, wuji video, …
Discover drivers, datasets, and MCP servers with: wuji list
Inspect details with: wuji info
Use --driver to override the backend for one run, or set per-capability defaults in .wuji/config.yaml.
Fine-tune models with: wuji train <capability> [method] [name] --dataset …

Memory (RAM/VRAM) is managed automatically by the core daemon.
Preload or free models with: wuji load, wuji unload, wuji kill
See: wuji resources explain`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version.Version, version.Commit, version.BuildDate),
	}

	root.PersistentFlags().StringP("driver", "d", "", "driver/backend to use (default: from config or dummy)")

	root.AddCommand(
		newListCmd(app),
		newInfoCmd(app),
		newLoadCmd(app),
		newUnloadCmd(app),
		newKillCmd(app),
		newTextCmd(app),
		newImageCmd(app),
		newVideoCmd(app),
		newAudioCmd(app),
		newMeshCmd(app),
		newDatasetCmd(app),
		newDriverCmd(app),
		newMcpCmd(app),
		newConfigCmd(app),
		newStatsCmd(app),
		newResourcesCmd(app),
		newGenerateCmd(app),
		newTrainCmd(app),
		newDataCmd(app),
		newRAGCmd(app),
		newCatalogCmd(app),
		newProviderCmd(app),
	)

	app.Root = root
	return app, nil
}

func isLocalCommand(args []string) bool {
	top := commandPath(args)
	if len(top) == 0 {
		return false
	}
	switch top[0] {
	case "catalog", "provider", "providers", "auth":
		return true
	case "config":
		return true
	case "list":
		if len(top) >= 2 {
			switch top[1] {
			case "providers", "provider", "models", "model", "labs", "lab", "routes", "route":
				return true
			}
		}
	}
	return false
}

func commandPath(args []string) []string {
	var path []string
	for i, arg := range args {
		if i == 0 {
			continue
		}
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		path = append(path, arg)
	}
	return path
}

// Execute runs the CLI and handles cleanup.
func (a *App) Execute() {
	defer func() {
		if a.cleanup != nil {
			a.cleanup()
		}
	}()

	if err := a.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
