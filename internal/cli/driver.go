package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji/internal/capability"
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
	connectCmd.Flags().Bool("save", true, "persist endpoint in .wuji/drivers.yaml")

	cmd.AddCommand(
		connectCmd,
		&cobra.Command{
			Use:   "list",
			Short: "List registered drivers",
			RunE: func(cmd *cobra.Command, args []string) error {
				drivers := app.Core.ListDrivers()
				if len(drivers) == 0 {
					fmt.Println("No drivers registered.")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "ID\tNAME\tVERSION\tREMOTE\tCAPABILITIES")
				for _, d := range drivers {
					remote := "no"
					if d.Remote {
						remote = "yes"
					}

					caps := ""
					for i, c := range d.Capabilities {
						if i > 0 {
							caps += ", "
						}
						caps += c.String()
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.ID, d.Name, d.Version, remote, caps)
				}
				return w.Flush()
			},
		},
		&cobra.Command{
			Use:   "capabilities",
			Short: "List all supported capability types",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Supported capabilities:")
				for _, c := range capability.All() {
					fmt.Printf("  - %s\n", c)
				}
			},
		},
	)

	return cmd
}
