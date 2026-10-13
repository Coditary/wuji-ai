package cli

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func newDecideCmd(app *App) *cobra.Command {
	var (
		model         string
		questionsFile string
		imagePath     string
		useStdin      bool
		prettyJSON bool
		recordset  bool
	)

	cmd := &cobra.Command{
		Use:   "decide [text]",
		Short: "Typed decisions: choice, score, noul",
		Long: `Run structured decision questions through a decision backend (e.g. Laya).

Provide questions via --questions JSON file or stdin. The file may contain either:
  - a full payload: {"state": {...}, "questions": {...}}
  - or a questions-only map compatible with the Jev / Laya /v1/systemone API

Examples:
  wuji decide "Package arrived broken" --questions triage.json
  wuji decide --questions questions.json --stdin
  wuji decide "photo of damage" --image photo.jpg --questions vision.json --model laya-multilingual`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			req := driver.DecisionRequest{
				Model:         model,
				QuestionsFile: questionsFile,
				ImagePath:     imagePath,
				UseStdin:      useStdin,
			}
			if len(args) > 0 {
				req.Text = args[0]
			}
			if err := req.ResolveQuestions(); err != nil {
				return err
			}
			if err := req.Validate(); err != nil {
				return err
			}

			resp, err := app.Core.ProduceDecision(context.Background(), app.resolveDriver(cmd, capability.Decision), req)
			if err != nil {
				return err
			}

			if recordset {
				shape, err := driver.DecisionResponseToRecordSet(resp, data.Meta{
					Capability: string(capability.Decision),
					Driver:     resp.Driver,
					Model:      resp.Model,
				})
				if err != nil {
					return err
				}
				opts := data.DefaultExportOpts()
				opts.Format = data.FormatJSON
				opts.PrettyJSON = prettyJSON
				return shape.Export(os.Stdout, opts)
			}
			var raw []byte
			if prettyJSON {
				raw, err = json.MarshalIndent(resp, "", "  ")
			} else {
				raw, err = json.Marshal(resp)
			}
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(raw)
			if err == nil {
				_, err = os.Stdout.Write([]byte("\n"))
			}
			return err
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "model alias (driver-specific)")
	cmd.Flags().StringVar(&questionsFile, "questions", "", "JSON file with decision questions")
	cmd.Flags().StringVar(&imagePath, "image", "", "image path for vision-backed decision state")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "read questions JSON from stdin")
	cmd.Flags().BoolVar(&prettyJSON, "pretty", false, "pretty-print JSON output")
	cmd.Flags().BoolVar(&recordset, "recordset", false, "emit WDD recordset instead of raw decision response")

	return cmd
}
