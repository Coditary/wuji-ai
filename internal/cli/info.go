package cli

import (
	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/mcp"
)

func newInfoCmd(app *App) *cobra.Command {
	mgr := mcp.NewManager(app.Config.Root)

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show detailed information about drivers, MCP servers, config, and collections",
		Long: `Inspect configuration and component details.

  wuji info driver <id>         driver capabilities, tasks, and formats
  wuji info mcp <name>          MCP server configuration and status
  wuji info config              effective .wuji/config.yaml
  wuji info collection <name>   RAG collection metadata`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printInfoOverview()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:     "driver [driver-id]",
			Aliases: []string{"drivers"},
			Short:   "Show driver details including supported source formats",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return printDriverInfo(app, args[0])
			},
		},
		&cobra.Command{
			Use:   "mcp [name]",
			Short: "Show detailed status for one MCP server (lists all when name is omitted)",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return printMCPList(cmd.Context(), mgr)
				}
				st, err := mgr.Status(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				return printMCPDetail(st)
			},
		},
		&cobra.Command{
			Use:   "config",
			Short: "Show effective configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printEffectiveConfig(app.Config)
			},
		},
		&cobra.Command{
			Use:     "collection [name]",
			Aliases: []string{"collections"},
			Short:   "Show RAG collection metadata",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return printRAGCollectionInfo(app, cmd, args[0])
			},
		},
	)

	return cmd
}
