package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/metrics"
)

func newStatsCmd(app *App) *cobra.Command {
	var since time.Duration
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show text generation usage statistics",
		Long:  "Summarize logged text generation metrics from .wuji/metrics/requests.jsonl.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cutoff := time.Time{}
			window := "all time"
			if since > 0 {
				cutoff = time.Now().UTC().Add(-since)
				window = since.String()
			}
			records, err := metrics.ReadSince(app.Config.Root, cutoff)
			if err != nil {
				return err
			}
			s := metrics.Summarize(records)
			fmt.Print(metrics.FormatSummary(s, window))
			return nil
		},
	}
	cmd.Flags().DurationVar(&since, "since", 0, "only include records within this duration (e.g. 24h, 7d)")
	return cmd
}
