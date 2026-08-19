package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji/internal/core"
	"github.com/coditary/wuji/pkg/version"
)

// App holds the CLI application state.
type App struct {
	Core *core.Core
	Root *cobra.Command
}

// New creates the root CLI application.
func New() (*App, error) {
	c, err := core.New(core.Config{})
	if err != nil {
		return nil, fmt.Errorf("initialize core: %w", err)
	}

	app := &App{Core: c}

	root := &cobra.Command{
		Use:   "wuji",
		Short: "Unified CLI for AI backends and APIs",
		Long: `Wuji is a unified command-line interface for AI workloads.

It provides a pluggable driver architecture where backends can be swapped
without changing how you interact with the tool.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version.Version, version.Commit, version.BuildDate),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.AddCommand(
		newGenerateCmd(app),
		newTTSCmd(app),
		newSTTCmd(app),
		newVoiceCmd(app),
		newTrainCmd(app),
		newDatasetCmd(app),
		newDriverCmd(app),
		newServeCmd(app),
	)

	app.Root = root
	return app, nil
}

// Execute runs the CLI and handles cleanup.
func (a *App) Execute() {
	defer func() {
		if a.Core != nil {
			_ = a.Core.Close()
		}
	}()

	if err := a.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
