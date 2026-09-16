package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newDriverCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "driver",
		Short: "Manage AI backends (drivers)",
	}

	connectCmd := &cobra.Command{
		Use:   "connect [endpoint]",
		Short: "Connect to a remote driver via gRPC",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			persist, _ := cmd.Flags().GetBool("save")
			if err := app.Core.ConnectRemote(cmd.Context(), args[0], persist); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Connected to driver at %s\n", args[0])
			return nil
		},
	}
	connectCmd.Flags().Bool("save", true, "persist endpoint in .wuji/config.yaml (driver_endpoints)")

	cmd.AddCommand(
		connectCmd,
		&cobra.Command{
			Use:        "list",
			Short:      "List registered drivers",
			Deprecated: "use `wuji list drivers` instead",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printDriverList(app)
			},
		},
		&cobra.Command{
			Use:        "info [driver-id]",
			Short:      "Show driver details including supported source formats",
			Deprecated: "use `wuji info driver` instead",
			Args:       cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return printDriverInfo(app, args[0])
			},
		},
		&cobra.Command{
			Use:        "formats",
			Short:      "List known model/source format identifiers",
			Deprecated: "use `wuji list formats` instead",
			Run: func(cmd *cobra.Command, args []string) {
				printFormatList()
			},
		},
		&cobra.Command{
			Use:        "capabilities",
			Short:      "List all supported capability types",
			Deprecated: "use `wuji list capabilities` instead",
			Run: func(cmd *cobra.Command, args []string) {
				printCapabilityList()
			},
		},
	)

	return cmd
}
