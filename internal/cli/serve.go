package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	grpcserver "github.com/coditary/wuji/internal/server/grpc"
)

func newServeCmd(app *App) *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Wuji core gRPC server for remote drivers",
		Long: `Start the Wuji core gRPC server.

Remote drivers connect to this server to register themselves
and expose their capabilities to the CLI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv, err := grpcserver.NewServer(app.Core, addr)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Wuji gRPC server listening on %s\n", srv.Addr())

			errCh := make(chan error, 1)
			go func() {
				errCh <- srv.Start()
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			select {
			case err := <-errCh:
				return err
			case <-sigCh:
				fmt.Fprintln(os.Stderr, "Shutting down...")
				srv.Stop()
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:50051", "gRPC listen address")
	return cmd
}
