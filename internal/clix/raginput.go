package clix

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coditary/wuji-core/pkg/driver"
)

type RAGRequestFields struct {
	Args          []string
	StoreRoot     string
	StoreDir      string
	Collection    string
	DirPaths      []string
	FilePaths     []string
	UseStdin      bool
	Recursive     bool
	ChunkSize     int
	ChunkOverlap  int
	EmbedModel    string
	TextModel     string
	SystemPrompt  string
	TopK          int
	MinScore      float32
	MaxTokens     int
	Filter        map[string]string
	PurgeSource   string
	IndexMode     string
	Force         bool
	DryRun        bool
	IndexMetadata map[string]string
	Glob          string
	Exclude       []string
	URL           string
	QueryFile     string
	QueryStdin    bool
	RerankTop     int
	Diverse       bool
	ContextMax    int
	Cite          bool
	Temperature   float32
	TopP          float32
	NoContext     bool
	IncludeScores bool
	RenameTo      string
	ExportPath    string
	ImportPath    string
	PreviewLength int
	QueryStdinR   io.Reader
}

func BuildRAGRequest(fields RAGRequestFields, task driver.RAGTask) (driver.RAGRequest, error) {
	query, err := resolveRAGQuery(fields)
	if err != nil {
		return driver.RAGRequest{}, err
	}
	paths := append(append([]string{}, fields.DirPaths...), fields.FilePaths...)
	mode := driver.RAGIndexMode(fields.IndexMode)
	if mode == "" {
		mode = driver.RAGIndexReplace
	}
	return driver.RAGRequest{
		Task: task, StoreRoot: fields.StoreRoot, StoreDir: fields.StoreDir, Collection: fields.Collection, Query: query,
		SourcePaths: paths, UseStdin: fields.UseStdin, Recursive: fields.Recursive,
		ChunkSize: fields.ChunkSize, ChunkOverlap: fields.ChunkOverlap, EmbedModel: fields.EmbedModel,
		TextModel: fields.TextModel, SystemPrompt: fields.SystemPrompt, TopK: fields.TopK,
		MinScore: fields.MinScore, MaxTokens: fields.MaxTokens, Filter: fields.Filter,
		PurgeSource: fields.PurgeSource, IndexMode: mode, Force: fields.Force, DryRun: fields.DryRun,
		IndexMetadata: fields.IndexMetadata, Glob: fields.Glob, Exclude: fields.Exclude, URL: fields.URL,
		QueryFile: fields.QueryFile, QueryStdin: fields.QueryStdin, RerankTop: fields.RerankTop,
		Diverse: fields.Diverse, ContextMaxChars: fields.ContextMax, Cite: fields.Cite,
		Temperature: fields.Temperature, TopP: fields.TopP, NoContext: fields.NoContext,
		IncludeScores: fields.IncludeScores, RenameTo: fields.RenameTo,
		ExportPath: fields.ExportPath, ImportPath: fields.ImportPath,
	}, nil
}

func resolveRAGQuery(fields RAGRequestFields) (string, error) {
	sources := 0
	query := strings.TrimSpace(strings.Join(fields.Args, " "))
	if query != "" {
		sources++
	}
	if fields.QueryFile != "" {
		sources++
	}
	if fields.QueryStdin {
		sources++
	}
	if sources > 1 {
		return "", fmt.Errorf("use one query source: positional text, --text/--query-file, or --query-stdin")
	}
	if fields.QueryFile != "" {
		data, err := os.ReadFile(fields.QueryFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	if fields.QueryStdin {
		r := fields.QueryStdinR
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
