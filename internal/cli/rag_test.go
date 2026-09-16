package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/coditary/wuji-core/pkg/ragstore"
)

func TestRAGIndexAndQueryDummy(t *testing.T) {
	app := newTestApp(t)
	root := t.TempDir()
	doc := filepath.Join(root, "doc.txt")
	if err := os.WriteFile(doc, []byte("Wuji CLI supports structured data and RAG retrieval."), 0o644); err != nil {
		t.Fatal(err)
	}

	driverID := app.Core.DefaultDriverID()
	indexShape, err := app.Core.RunRAG(context.Background(), driverID, driver.RAGRequest{
		Task: driver.RAGTaskIndex, StoreRoot: root, Collection: "test",
		SourcePaths: []string{doc}, ChunkSize: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if indexShape.Kind() != data.ShapeRecordSet {
		t.Fatalf("index kind=%s", indexShape.Kind())
	}

	queryShape, err := app.Core.RunRAG(context.Background(), driverID, driver.RAGRequest{
		Task: driver.RAGTaskQuery, StoreRoot: root, Collection: "test",
		Query: "structured data", TopK: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := queryShape.Export(&buf, data.DefaultExportOpts()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "chunk") {
		t.Fatalf("query output: %s", buf.String())
	}
}

func TestRAGIndexAndQueryLocalRAG(t *testing.T) {
	app := newTestApp(t)
	root := t.TempDir()
	doc := filepath.Join(root, "local.txt")
	if err := os.WriteFile(doc, []byte("Local driver indexes with data embeddings."), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := app.Core.RunRAG(context.Background(), "local-rag", driver.RAGRequest{
		Task: driver.RAGTaskIndex, StoreRoot: root, Collection: "local-rag",
		SourcePaths: []string{doc}, ChunkSize: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	col, err := ragstore.Load(context.Background(), root, "local-rag")
	if err != nil {
		t.Fatal(err)
	}
	if col.Manifest.Dims != ragstore.DefaultEmbedDims() {
		t.Fatalf("expected pseudo embed dims=%d, got %d", ragstore.DefaultEmbedDims(), col.Manifest.Dims)
	}
}
