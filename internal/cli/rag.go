package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-ai/internal/clix"
	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/data"
)

func newRAGCmd(app *App) *cobra.Command {
	var fields clix.RAGRequestFields
	var (
		jsonOut, csvOut, ndjsonOut, msgpackOut bool
		format, outputFormat, outputShape      string
		queryTextFile                          string
		appendIndex, replaceIndex              bool
		prettyJSON, quiet, verbose, noScores   bool
		csvFields                              string
		opFlags                                clix.RAGOpFlags
	)

	cmd := &cobra.Command{
		Use:   "rag [query]",
		Short: "Retrieval-augmented generation: index, query, answer",
		Long: `Index documents, retrieve relevant chunks, and optionally generate answers.

Operations (pick one):
  --index              ingest --dir/--file/--stdin into --collection
  --query              retrieve top chunks for a query
  --answer             retrieve chunks and generate an answer via wuji text
  --list-collections   list local collections (prefer: wuji list collections)
  --info               show collection metadata (prefer: wuji info collection <name>)
  --stats              show detailed collection statistics
  --delete             delete a collection
  --purge              remove one --source from a collection
  --export             copy collection to --to
  --import             load collection from --from
  --rename             rename --collection to --rename-to

Examples:
  wuji rag --index --dir ./docs --collection handbook --glob "*.md"
  wuji rag "Kündigungsfrist?" --query --collection handbook --store-root . --top 5
  wuji rag "Frage?" --query --collection default --store-root . --store-dir .taiji
  wuji rag "Frage?" --answer --collection handbook --cite --quiet
  wuji rag --stats --collection handbook --verbose

Output uses WDD (default JSON). See also: --json, --csv, --msgpack, --format.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := opFlags.Resolve()
			if err != nil {
				return err
			}
			fields.Args = args
			if queryTextFile != "" && fields.QueryFile == "" {
				fields.QueryFile = queryTextFile
			}
			if appendIndex {
				fields.IndexMode = "append"
			} else if replaceIndex {
				fields.IndexMode = "replace"
			}
			if fields.StoreRoot == "" {
				fields.StoreRoot = app.Config.Root
			}
			if fields.StoreRoot == "" {
				fields.StoreRoot = "."
			}
			if !fields.IncludeScores && !noScores {
				fields.IncludeScores = true
			}
			if noScores {
				fields.IncludeScores = false
			}

			req, err := clix.BuildRAGRequest(fields, task)
			if err != nil {
				return err
			}

			exportFmt, err := clix.ParseExportFormat(jsonOut, csvOut, ndjsonOut, msgpackOut, format)
			if err != nil {
				return err
			}
			if outputFormat != "" {
				exportFmt = data.ExportFormat(outputFormat)
			}

			shape, err := app.Core.RunRAG(context.Background(), app.resolveDriver(cmd, capability.RAG), req)
			if err != nil {
				return err
			}
			shape = clix.ApplyRAGOutputShape(shape, outputShape)

			opts := data.DefaultExportOpts()
			opts.Format = exportFmt
			opts.PrettyJSON = prettyJSON
			opts.CSVFields = clix.ParseCSVFields(csvFields)
			opts.PreviewLength = fields.PreviewLength

			if verbose {
				clix.WriteRAGVerbose(os.Stderr, shape, task)
			}
			return clix.WriteRAGOutput(shape, opts, quiet, task)
		},
	}

	clix.AddRAGOpFlags(cmd, &opFlags)

	cmd.Flags().StringVar(&fields.StoreRoot, "store-root", "", "project root for RAG storage (default: config root)")
	cmd.Flags().StringVar(&fields.StoreDir, "store-dir", "", "state directory under store-root (default: .wuji; Taiji local index: .taiji)")
	cmd.Flags().StringVar(&fields.Collection, "collection", "", "collection name")
	cmd.Flags().StringSliceVar(&fields.DirPaths, "dir", nil, "directory to index (repeatable)")
	cmd.Flags().StringSliceVar(&fields.FilePaths, "file", nil, "file to index (repeatable)")
	cmd.Flags().BoolVar(&fields.UseStdin, "stdin", false, "read index input from stdin")
	cmd.Flags().BoolVar(&fields.Recursive, "recursive", true, "recurse into --dir")
	cmd.Flags().IntVar(&fields.ChunkSize, "chunk-size", 800, "chunk size in runes for --index")
	cmd.Flags().IntVar(&fields.ChunkOverlap, "chunk-overlap", 100, "chunk overlap in runes for --index")
	cmd.Flags().StringVar(&fields.EmbedModel, "embed-model", "", "embedding model (driver-specific)")
	cmd.Flags().StringVar(&fields.IndexMode, "index-mode", "replace", "index merge mode: replace, append")
	cmd.Flags().BoolVar(&appendIndex, "append", false, "append chunks without replacing existing sources")
	cmd.Flags().BoolVar(&replaceIndex, "replace", false, "replace chunks for indexed sources (default)")
	cmd.Flags().BoolVar(&fields.Force, "force", false, "index despite embed model/dims mismatch")
	cmd.Flags().BoolVar(&fields.DryRun, "dry-run", false, "show index plan without saving")
	cmd.Flags().StringToStringVar(&fields.IndexMetadata, "metadata", nil, "metadata key=value attached to indexed chunks")
	cmd.Flags().StringVar(&fields.Glob, "glob", "", "filename glob for files under --dir (e.g. *.md)")
	cmd.Flags().StringSliceVar(&fields.Exclude, "exclude", nil, "skip paths or patterns during indexing")
	cmd.Flags().StringVar(&fields.URL, "url", "", "URL to index (not implemented)")

	cmd.Flags().IntVar(&fields.TopK, "top", 5, "number of chunks to retrieve")
	cmd.Flags().Float32Var(&fields.MinScore, "min-score", 0, "minimum similarity score")
	cmd.Flags().StringToStringVar(&fields.Filter, "filter", nil, "metadata filter key=value for retrieval")
	cmd.Flags().StringVar(&fields.QueryFile, "query-file", "", "read query text from file")
	cmd.Flags().StringVar(&queryTextFile, "text", "", "read query text from file (alias for --query-file)")
	cmd.Flags().BoolVar(&fields.QueryStdin, "query-stdin", false, "read query text from stdin")
	cmd.Flags().IntVar(&fields.RerankTop, "rerank-top", 0, "retrieve N candidates before trimming to --top")
	cmd.Flags().BoolVar(&fields.Diverse, "diverse", false, "apply MMR diversification when ranking chunks")
	cmd.Flags().BoolVar(&noScores, "no-scores", false, "omit score fields from chunk output")

	cmd.Flags().StringVar(&fields.TextModel, "text-model", "", "text model for --answer")
	cmd.Flags().StringVar(&fields.SystemPrompt, "system-prompt", "", "extra system prompt for --answer")
	cmd.Flags().IntVar(&fields.MaxTokens, "max-tokens", 1024, "max tokens for --answer")
	cmd.Flags().IntVar(&fields.ContextMax, "context-max-chars", 0, "truncate retrieved context for --answer")
	cmd.Flags().BoolVar(&fields.Cite, "cite", false, "ask the model to cite chunks as [1], [2], …")
	cmd.Flags().Float32Var(&fields.Temperature, "temperature", 0, "sampling temperature for --answer")
	cmd.Flags().Float32Var(&fields.TopP, "top-p", 0, "nucleus sampling for --answer")
	cmd.Flags().BoolVar(&fields.NoContext, "no-context", false, "omit retrieved chunks from --answer output")

	cmd.Flags().StringVar(&fields.PurgeSource, "source", "", "source path to purge from collection")
	cmd.Flags().StringVar(&fields.ExportPath, "to", "", "destination path for --export")
	cmd.Flags().StringVar(&fields.ImportPath, "from", "", "source path for --import")
	cmd.Flags().StringVar(&fields.RenameTo, "rename-to", "", "new collection name for --rename")

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit WDD JSON")
	cmd.Flags().BoolVar(&csvOut, "csv", false, "emit CSV projection")
	cmd.Flags().BoolVar(&ndjsonOut, "ndjson", false, "emit NDJSON records")
	cmd.Flags().BoolVar(&msgpackOut, "msgpack", false, "emit MessagePack")
	cmd.Flags().StringVar(&format, "format", "", "output format: json, csv, ndjson, msgpack")
	cmd.Flags().StringVar(&outputFormat, "output-format", "", "alias for --format")
	cmd.Flags().StringVar(&outputShape, "output-shape", "", "output projection: full or answer")
	cmd.Flags().StringVar(&csvFields, "fields", "", "CSV columns (comma-separated, e.g. text,source,score)")
	cmd.Flags().IntVar(&fields.PreviewLength, "preview-length", 0, "truncate long text fields in CSV output")
	cmd.Flags().BoolVar(&prettyJSON, "pretty", true, "pretty-print JSON output")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "with --answer, print answer text only")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "print operation summary on stderr")

	return cmd
}
