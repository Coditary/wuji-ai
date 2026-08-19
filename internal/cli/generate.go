package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji/internal/driver"
)

func newGenerateCmd(app *App) *cobra.Command {
	var (
		driverID    string
		maxTokens   int
		temperature float32
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate content using AI backends",
	}

	textCmd := &cobra.Command{
		Use:   "text [prompt]",
		Short: "Generate text from a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]
			if len(args) > 1 {
				for _, part := range args[1:] {
					prompt += " " + part
				}
			}

			resp, err := app.Core.GenerateText(context.Background(), driverID, driver.TextRequest{
				Prompt:      prompt,
				MaxTokens:   maxTokens,
				Temperature: temperature,
			})
			if err != nil {
				return err
			}

			fmt.Fprintln(os.Stdout, resp.Text)
			return nil
		},
	}

	textCmd.Flags().StringVarP(&driverID, "driver", "d", "", "driver to use (default: configured default)")
	textCmd.Flags().IntVar(&maxTokens, "max-tokens", 256, "maximum tokens to generate")
	textCmd.Flags().Float32Var(&temperature, "temperature", 0.7, "sampling temperature")

	cmd.AddCommand(textCmd)
	return cmd
}
