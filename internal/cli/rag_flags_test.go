package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildRAGRequestDryRun(t *testing.T) {
	req, err := buildRAGRequest(ragRequestFields{
		storeRoot: t.TempDir(), collection: "x", dirPaths: []string{"."},
		indexMode: "replace", dryRun: true,
	}, driver.RAGTaskIndex)
	if err != nil {
		t.Fatal(err)
	}
	if !req.DryRun || req.Collection != "x" {
		t.Fatalf("req=%+v", req)
	}
}

func TestBuildRAGRequestFilter(t *testing.T) {
	req, err := buildRAGRequest(ragRequestFields{
		storeRoot: ".", collection: "x", args: []string{"q"},
		filter: map[string]string{"source": "a.txt"},
	}, driver.RAGTaskQuery)
	if err != nil {
		t.Fatal(err)
	}
	if req.Filter["source"] != "a.txt" {
		t.Fatalf("filter=%v", req.Filter)
	}
}

func TestRAGDryRunIndex(t *testing.T) {
	app := newTestApp(t)
	root := t.TempDir()
	doc := filepath.Join(root, "doc.txt")
	if err := os.WriteFile(doc, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	shape, err := app.Core.RunRAG(context.Background(), app.Core.DefaultDriverID(), driver.RAGRequest{
		Task: driver.RAGTaskIndex, StoreRoot: root, Collection: "dry",
		SourcePaths: []string{doc}, DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	rs, ok := shape.(*data.RecordSet)
	if !ok {
		t.Fatalf("got %T", shape)
	}
	for _, rec := range rs.Records {
		if rec.Role == "result" {
			if v, ok := rec.Fields["dry_run"]; ok && v.B {
				return
			}
		}
	}
	t.Fatalf("expected dry_run result: %+v", rs.Records)
}
