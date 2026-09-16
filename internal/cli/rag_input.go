package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coditary/wuji-core/pkg/driver"
)

type ragRequestFields struct {
	args           []string
	storeRoot      string
	storeDir       string
	collection     string
	dirPaths       []string
	filePaths      []string
	useStdin       bool
	recursive      bool
	chunkSize      int
	chunkOverlap   int
	embedModel     string
	textModel      string
	systemPrompt   string
	topK           int
	minScore       float32
	maxTokens      int
	filter         map[string]string
	purgeSource    string
	indexMode      string
	force          bool
	dryRun         bool
	indexMetadata  map[string]string
	glob           string
	exclude        []string
	url            string
	queryFile      string
	queryStdin     bool
	rerankTop      int
	diverse        bool
	contextMax     int
	cite           bool
	temperature    float32
	topP           float32
	noContext      bool
	includeScores  bool
	renameTo       string
	exportPath     string
	importPath     string
	previewLength  int
	queryStdinR    io.Reader
}

func buildRAGRequest(fields ragRequestFields, task driver.RAGTask) (driver.RAGRequest, error) {
	query, err := resolveRAGQuery(fields)
	if err != nil {
		return driver.RAGRequest{}, err
	}
	paths := append(append([]string{}, fields.dirPaths...), fields.filePaths...)
	mode := driver.RAGIndexMode(fields.indexMode)
	if mode == "" {
		mode = driver.RAGIndexReplace
	}
	return driver.RAGRequest{
		Task: task, StoreRoot: fields.storeRoot, StoreDir: fields.storeDir, Collection: fields.collection, Query: query,
		SourcePaths: paths, UseStdin: fields.useStdin, Recursive: fields.recursive,
		ChunkSize: fields.chunkSize, ChunkOverlap: fields.chunkOverlap, EmbedModel: fields.embedModel,
		TextModel: fields.textModel, SystemPrompt: fields.systemPrompt, TopK: fields.topK,
		MinScore: fields.minScore, MaxTokens: fields.maxTokens, Filter: fields.filter,
		PurgeSource: fields.purgeSource, IndexMode: mode, Force: fields.force, DryRun: fields.dryRun,
		IndexMetadata: fields.indexMetadata, Glob: fields.glob, Exclude: fields.exclude, URL: fields.url,
		QueryFile: fields.queryFile, QueryStdin: fields.queryStdin, RerankTop: fields.rerankTop,
		Diverse: fields.diverse, ContextMaxChars: fields.contextMax, Cite: fields.cite,
		Temperature: fields.temperature, TopP: fields.topP, NoContext: fields.noContext,
		IncludeScores: fields.includeScores, RenameTo: fields.renameTo,
		ExportPath: fields.exportPath, ImportPath: fields.importPath,
	}, nil
}

func resolveRAGQuery(fields ragRequestFields) (string, error) {
	sources := 0
	query := strings.TrimSpace(strings.Join(fields.args, " "))
	if query != "" {
		sources++
	}
	if fields.queryFile != "" {
		sources++
	}
	if fields.queryStdin {
		sources++
	}
	if sources > 1 {
		return "", fmt.Errorf("use one query source: positional text, --text/--query-file, or --query-stdin")
	}
	if fields.queryFile != "" {
		data, err := os.ReadFile(fields.queryFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	if fields.queryStdin {
		r := fields.queryStdinR
		if r == nil {
			r = os.Stdin
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	return query, nil
}
