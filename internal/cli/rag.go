package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/data"
)

func newRAGCmd(app *App) *cobra.Command {
	var fields ragRequestFields
	var (
		jsonOut, csvOut, ndjsonOut, msgpackOut bool
		format, outputFormat, outputShape      string
		queryTextFile                          string
		appendIndex, replaceIndex              bool
		prettyJSON, quiet, verbose, noScores   bool
		csvFields                              string
		opFlags                                ragOpFlags
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
			task, err := opFlags.resolve()
			if err != nil {
				return err
			}
			fields.args = args
			if queryTextFile != "" && fields.queryFile == "" {
				fields.queryFile = queryTextFile
			}
			if appendIndex {
				fields.indexMode = "append"
			} else if replaceIndex {
				fields.indexMode = "replace"
			}
			if fields.storeRoot == "" {
				fields.storeRoot = app.Config.Root
			}
			if fields.storeRoot == "" {
				fields.storeRoot = "."
			}
			if !fields.includeScores && !noScores {
				fields.includeScores = true
			}
			if noScores {
				fields.includeScores = false
			}

			req, err := buildRAGRequest(fields, task)
			if err != nil {
				return err
			}

			exportFmt, err := parseExportFormat(jsonOut, csvOut, ndjsonOut, msgpackOut, format)
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
			shape = applyRAGOutputShape(shape, outputShape)

			opts := data.DefaultExportOpts()
			opts.Format = exportFmt
			opts.PrettyJSON = prettyJSON
			opts.CSVFields = parseCSVFields(csvFields)
			opts.PreviewLength = fields.previewLength

			if verbose {
				writeRAGVerbose(os.Stderr, shape, task)
			}
			return writeRAGOutput(shape, opts, quiet, task)
		},
	}

	addRAGOpFlags(cmd, &opFlags)

	cmd.Flags().StringVar(&fields.storeRoot, "store-root", "", "project root for RAG storage (default: config root)")
	cmd.Flags().StringVar(&fields.storeDir, "store-dir", "", "state directory under store-root (default: .wuji; Taiji local index: .taiji)")
	cmd.Flags().StringVar(&fields.collection, "collection", "", "collection name")
	cmd.Flags().StringSliceVar(&fields.dirPaths, "dir", nil, "directory to index (repeatable)")
	cmd.Flags().StringSliceVar(&fields.filePaths, "file", nil, "file to index (repeatable)")
	cmd.Flags().BoolVar(&fields.useStdin, "stdin", false, "read index input from stdin")
	cmd.Flags().BoolVar(&fields.recursive, "recursive", true, "recurse into --dir")
	cmd.Flags().IntVar(&fields.chunkSize, "chunk-size", 800, "chunk size in runes for --index")
	cmd.Flags().IntVar(&fields.chunkOverlap, "chunk-overlap", 100, "chunk overlap in runes for --index")
	cmd.Flags().StringVar(&fields.embedModel, "embed-model", "", "embedding model (driver-specific)")
	cmd.Flags().StringVar(&fields.indexMode, "index-mode", "replace", "index merge mode: replace, append")
	cmd.Flags().BoolVar(&appendIndex, "append", false, "append chunks without replacing existing sources")
	cmd.Flags().BoolVar(&replaceIndex, "replace", false, "replace chunks for indexed sources (default)")
	cmd.Flags().BoolVar(&fields.force, "force", false, "index despite embed model/dims mismatch")
	cmd.Flags().BoolVar(&fields.dryRun, "dry-run", false, "show index plan without saving")
	cmd.Flags().StringToStringVar(&fields.indexMetadata, "metadata", nil, "metadata key=value attached to indexed chunks")
	cmd.Flags().StringVar(&fields.glob, "glob", "", "filename glob for files under --dir (e.g. *.md)")
	cmd.Flags().StringSliceVar(&fields.exclude, "exclude", nil, "skip paths or patterns during indexing")
	cmd.Flags().StringVar(&fields.url, "url", "", "URL to index (not implemented)")

	cmd.Flags().IntVar(&fields.topK, "top", 5, "number of chunks to retrieve")
	cmd.Flags().Float32Var(&fields.minScore, "min-score", 0, "minimum similarity score")
	cmd.Flags().StringToStringVar(&fields.filter, "filter", nil, "metadata filter key=value for retrieval")
	cmd.Flags().StringVar(&fields.queryFile, "query-file", "", "read query text from file")
	cmd.Flags().StringVar(&queryTextFile, "text", "", "read query text from file (alias for --query-file)")
	cmd.Flags().BoolVar(&fields.queryStdin, "query-stdin", false, "read query text from stdin")
	cmd.Flags().IntVar(&fields.rerankTop, "rerank-top", 0, "retrieve N candidates before trimming to --top")
	cmd.Flags().BoolVar(&fields.diverse, "diverse", false, "apply MMR diversification when ranking chunks")
	cmd.Flags().BoolVar(&noScores, "no-scores", false, "omit score fields from chunk output")

	cmd.Flags().StringVar(&fields.textModel, "text-model", "", "text model for --answer")
	cmd.Flags().StringVar(&fields.systemPrompt, "system-prompt", "", "extra system prompt for --answer")
	cmd.Flags().IntVar(&fields.maxTokens, "max-tokens", 1024, "max tokens for --answer")
	cmd.Flags().IntVar(&fields.contextMax, "context-max-chars", 0, "truncate retrieved context for --answer")
	cmd.Flags().BoolVar(&fields.cite, "cite", false, "ask the model to cite chunks as [1], [2], …")
	cmd.Flags().Float32Var(&fields.temperature, "temperature", 0, "sampling temperature for --answer")
	cmd.Flags().Float32Var(&fields.topP, "top-p", 0, "nucleus sampling for --answer")
	cmd.Flags().BoolVar(&fields.noContext, "no-context", false, "omit retrieved chunks from --answer output")

	cmd.Flags().StringVar(&fields.purgeSource, "source", "", "source path to purge from collection")
	cmd.Flags().StringVar(&fields.exportPath, "to", "", "destination path for --export")
	cmd.Flags().StringVar(&fields.importPath, "from", "", "source path for --import")
	cmd.Flags().StringVar(&fields.renameTo, "rename-to", "", "new collection name for --rename")

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit WDD JSON")
	cmd.Flags().BoolVar(&csvOut, "csv", false, "emit CSV projection")
	cmd.Flags().BoolVar(&ndjsonOut, "ndjson", false, "emit NDJSON records")
	cmd.Flags().BoolVar(&msgpackOut, "msgpack", false, "emit MessagePack")
	cmd.Flags().StringVar(&format, "format", "", "output format: json, csv, ndjson, msgpack")
	cmd.Flags().StringVar(&outputFormat, "output-format", "", "alias for --format")
	cmd.Flags().StringVar(&outputShape, "output-shape", "", "output projection: full or answer")
	cmd.Flags().StringVar(&csvFields, "fields", "", "CSV columns (comma-separated, e.g. text,source,score)")
	cmd.Flags().IntVar(&fields.previewLength, "preview-length", 0, "truncate long text fields in CSV output")
	cmd.Flags().BoolVar(&prettyJSON, "pretty", true, "pretty-print JSON output")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "with --answer, print answer text only")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "print operation summary on stderr")

	return cmd
}
