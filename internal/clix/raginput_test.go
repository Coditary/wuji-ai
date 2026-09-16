package clix

import (
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildRAGRequestDryRun(t *testing.T) {
	req, err := BuildRAGRequest(RAGRequestFields{
		StoreRoot: t.TempDir(), Collection: "x", DirPaths: []string{"."},
		IndexMode: "replace", DryRun: true,
	}, driver.RAGTaskIndex)
	if err != nil {
		t.Fatal(err)
	}
	if !req.DryRun || req.Collection != "x" {
		t.Fatalf("req=%+v", req)
	}
}

func TestBuildRAGRequestFilter(t *testing.T) {
	req, err := BuildRAGRequest(RAGRequestFields{
		StoreRoot: ".", Collection: "x", Args: []string{"q"},
		Filter: map[string]string{"source": "a.txt"},
	}, driver.RAGTaskQuery)
	if err != nil {
		t.Fatal(err)
	}
	if req.Filter["source"] != "a.txt" {
		t.Fatalf("filter=%v", req.Filter)
	}
}

func TestBuildRAGRequestQueryFromStdin(t *testing.T) {
	req, err := BuildRAGRequest(RAGRequestFields{
		QueryStdin: true, QueryStdinR: strings.NewReader("stdin query"),
	}, driver.RAGTaskQuery)
	if err != nil {
		t.Fatal(err)
	}
	if req.Query != "stdin query" {
		t.Fatalf("query=%q", req.Query)
	}
}

func TestBuildRAGRequestConflictingQuerySources(t *testing.T) {
	_, err := BuildRAGRequest(RAGRequestFields{
		Args: []string{"positional"}, QueryStdin: true,
	}, driver.RAGTaskQuery)
	if err == nil {
		t.Fatal("expected query source conflict")
	}
}

func TestRAGOpFlagsResolveRequired(t *testing.T) {
	_, err := RAGOpFlags{}.Resolve()
	if err == nil {
		t.Fatal("expected required operation flag error")
	}
}

func TestRAGOpFlagsResolveConflict(t *testing.T) {
	_, err := RAGOpFlags{Index: true, Query: true}.Resolve()
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestRAGOpFlagsResolveIndex(t *testing.T) {
	task, err := RAGOpFlags{Index: true}.Resolve()
	if err != nil || task != driver.RAGTaskIndex {
		t.Fatalf("task=%s err=%v", task, err)
	}
}
