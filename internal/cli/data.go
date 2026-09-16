package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func newDataCmd(app *App) *cobra.Command {
	var (
		model string
		textFile, imagePath, csvPath, graphPath, inputPath string
		inputFormat, outputFormat, format string
		inputShape, outputShape string
		useStdin bool
		jsonOut, csvOut, ndjsonOut, msgpackOut bool
		csvView, csvVector string
		prettyJSON bool
		horizon int
		taskFlags dataTaskFlags
	)

	cmd := &cobra.Command{
		Use:   "data [text]",
		Short: "Structured data: embed, vision, tables, graphs",
		Long: `Produce or transform structured machine-readable data.

Task flags (pick one):
  --embed, --vector       text/image embeddings
  --detect, --segment     computer vision (JSON recordset)
  --ner, --keypoint       token/spatial annotations
  --classify, --regress   tabular ML on --csv-file input
  --forecast              time series on --csv-file input (--horizon)
  --graph                 graph inference
  --node-classify         node classification on --graph-file
  --link-predict          link prediction on --graph-file

Examples:
  wuji data "Hallo" --embed
  wuji data --text doc.txt --ner
  wuji data --image photo.jpg --detect
  wuji data --csv-file sales.csv --forecast --horizon 7
  wuji data --graph-file net.json --node-classify

Convert-only (no task flag, reshape/format only):
  wuji data --input table.csv --input-format csv --format json
  wuji data --input doc.wdd --output-shape table --format csv

Output format (one of):
  --json (default)   WDD JSON document
  --csv              flattened projection
  --ndjson           one record per line
  --msgpack          binary MessagePack (compact vectors)
  --format json|csv|ndjson|msgpack

Shape control:
  --input-shape auto|recordset|table|vector|graph
  --output-shape auto|recordset|table|vector|graph

CSV options:
  --csv-view records|links|nodes
  --csv-vector json|expand|dims`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			exportFmt, err := parseExportFormat(jsonOut, csvOut, ndjsonOut, msgpackOut, format)
			if err != nil {
				return err
			}
			if outputFormat != "" {
				exportFmt = data.ExportFormat(outputFormat)
			}

			req, err := buildDataRequest(args, dataRequestFields{
				taskFlags: taskFlags, model: model, textFile: textFile, imagePath: imagePath,
				csvPath: csvPath, graphPath: graphPath, inputPath: inputPath,
				useStdin: useStdin, inputFormat: driver.DataInputFormat(inputFormat),
				inputShape: data.ShapeKind(inputShape), outputShape: data.ShapeKind(outputShape),
				horizon: horizon,
			})
			if err != nil {
				return err
			}

			shape, err := app.Core.ProduceData(context.Background(), app.resolveDriver(cmd, capability.Data), req)
			if err != nil {
				return err
			}

			opts := data.DefaultExportOpts()
			opts.Format = exportFmt
			opts.CSVView = data.CSVView(csvView)
			opts.VectorCSV = data.VectorCSVMode(csvVector)
			opts.PrettyJSON = prettyJSON
			if opts.CSVView == "" {
				opts.CSVView = data.CSVViewRecords
			}
			if opts.VectorCSV == "" {
				opts.VectorCSV = data.VectorCSVDims
			}
			return shape.Export(os.Stdout, opts)
		},
	}

	addDataTaskFlags(cmd, &taskFlags)
	cmd.Flags().StringVar(&model, "model", "", "model name (driver-specific)")
	cmd.Flags().StringVar(&textFile, "text", "", "read text input from file")
	cmd.Flags().StringVar(&imagePath, "image", "", "image file for vision tasks")
	cmd.Flags().StringVar(&csvPath, "csv-file", "", "CSV file for tabular tasks")
	cmd.Flags().StringVar(&graphPath, "graph-file", "", "graph JSON/WDD file for graph tasks")
	cmd.Flags().StringVar(&inputPath, "input", "", "generic input file (json or csv via --input-format)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "read structured input from stdin")
	cmd.Flags().StringVar(&inputFormat, "input-format", "auto", "input format: auto, json, csv, msgpack")
	cmd.Flags().StringVar(&inputShape, "input-shape", "auto", "interpret input as: auto, recordset, table, vector, graph")
	cmd.Flags().StringVar(&outputShape, "output-shape", "auto", "emit as shape: auto, recordset, table, vector, graph")
	cmd.Flags().IntVar(&horizon, "horizon", 7, "forecast horizon in steps (--forecast)")

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit WDD JSON (default when no format flag is set)")
	cmd.Flags().BoolVar(&csvOut, "csv", false, "emit CSV projection")
	cmd.Flags().BoolVar(&ndjsonOut, "ndjson", false, "emit newline-delimited JSON records")
	cmd.Flags().BoolVar(&msgpackOut, "msgpack", false, "emit MessagePack binary")
	cmd.Flags().StringVar(&format, "format", "", "output format: json, csv, ndjson, msgpack")
	cmd.Flags().StringVar(&outputFormat, "output-format", "", "alias for --format")
	cmd.Flags().BoolVar(&prettyJSON, "pretty", true, "pretty-print JSON output")

	cmd.Flags().StringVar(&csvView, "csv-view", "records", "CSV view: records, links, nodes")
	cmd.Flags().StringVar(&csvVector, "csv-vector", "dims", "vector CSV mode: json, expand, dims")

	return cmd
}
