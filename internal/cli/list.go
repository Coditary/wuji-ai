package cli

import (
	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/mcp"
)

func newListCmd(app *App) *cobra.Command {
	mgr := mcp.NewManager(app.Config.Root)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List drivers, models, MCP servers, datasets, and other catalogs",
		Long: `Discover what is available in your Wuji setup.

Run without a subcommand for an overview, or pick a topic:

  wuji list drivers
  wuji list mcp
  wuji list datasets
  wuji list loras
  wuji list providers
  wuji list models
  wuji list labs
  wuji list routes
  wuji list capabilities
  wuji list formats
  wuji list collections`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printListOverview(app)
		},
	}

	modelsCmd := &cobra.Command{
		Use:     "models",
		Aliases: []string{"model"},
		Short:   "List canonical models from the catalog cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			providerID, _ := cmd.Flags().GetString("provider")
			labID, _ := cmd.Flags().GetString("lab")
			return printCatalogModelList(app, providerID, labID)
		},
	}
	modelsCmd.Flags().String("provider", "", "list models for a catalog provider id (e.g. github-copilot)")
	modelsCmd.Flags().String("lab", "", "filter canonical models by lab id (e.g. google)")

	driversCmd := &cobra.Command{
		Use:     "drivers",
		Aliases: []string{"driver"},
		Short:   "List registered drivers",
		Long: `Lists built-in and configured plugin drivers.

Shows active drivers (registered with Wuji) and configured plugins with status
stopped, missing, or listening (gRPC up but not attached) without starting them.
Plugins start on first use (text, image, decide, …) when auto_start_driver is enabled.

Use --connect to attach or auto-start every configured plugin before listing
(default: false; use this if you relied on the old auto-connect behavior).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			connect, _ := cmd.Flags().GetBool("connect")
			if connect {
				connectConfiguredDrivers(cmd.Context(), app)
			}
			return printDriverList(app)
		},
	}
	driversCmd.Flags().Bool("connect", false, "attach or auto-start all configured plugin drivers before listing")

	cmd.AddCommand(
		driversCmd,
		&cobra.Command{
			Use:   "mcp",
			Short: "List configured MCP servers and their status",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printMCPList(cmd.Context(), mgr)
			},
		},
		&cobra.Command{
			Use:     "datasets",
			Aliases: []string{"dataset"},
			Short:   "List datasets",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printDatasetList(app, cmd)
			},
		},
		&cobra.Command{
			Use:     "loras",
			Aliases: []string{"lora"},
			Short:   "List configured LoRA aliases",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printLoRAList(app.Config)
			},
		},
		&cobra.Command{
			Use:     "providers",
			Aliases: []string{"provider"},
			Short:   "List models.dev API providers from the catalog cache",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printCatalogProviderList(app)
			},
		},
		modelsCmd,
		&cobra.Command{
			Use:     "labs",
			Aliases: []string{"lab"},
			Short:   "List model labs (authors) from the catalog cache",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printCatalogLabList(app)
			},
		},
		&cobra.Command{
			Use:     "routes",
			Aliases: []string{"route"},
			Short:   "List configured routes in .wuji/config.yaml",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printRouteList(app.Config)
			},
		},
		&cobra.Command{
			Use:   "capabilities",
			Short: "List all supported capability types",
			RunE: func(cmd *cobra.Command, args []string) error {
				printCapabilityList()
				return nil
			},
		},
		&cobra.Command{
			Use:   "formats",
			Short: "List known model/source format identifiers",
			RunE: func(cmd *cobra.Command, args []string) error {
				printFormatList()
				return nil
			},
		},
		&cobra.Command{
			Use:     "collections",
			Aliases: []string{"collection"},
			Short:   "List indexed RAG collections",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printRAGCollections(app, cmd)
			},
		},
	)

	return cmd
}
