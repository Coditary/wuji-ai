package clix

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func TestApplyRAGOutputShapeAnswerOnly(t *testing.T) {
	shape := &data.RecordSet{
		Records: []data.Record{
			{Role: "chunk", Fields: map[string]data.Value{"text": {Kind: data.KindString, S: "ctx"}}},
			{Role: "answer", Fields: map[string]data.Value{"text": {Kind: data.KindString, S: "ans"}}},
		},
	}
	out := ApplyRAGOutputShape(shape, string(driver.RAGOutputAnswerOnly))
	rs, ok := out.(*data.RecordSet)
	if !ok || len(rs.Records) != 1 || rs.Records[0].Role != "answer" {
		t.Fatalf("out=%+v", out)
	}
}

func TestParseCSVFields(t *testing.T) {
	fields := ParseCSVFields("text, source ,,score")
	if len(fields) != 3 || fields[0] != "text" || fields[2] != "score" {
		t.Fatalf("fields=%v", fields)
	}
	if ParseCSVFields("  ") != nil {
		t.Fatal("empty should be nil")
	}
}

func TestWriteRAGVerboseIndex(t *testing.T) {
	shape := &data.RecordSet{
		Records: []data.Record{{
			Role: "result",
			Fields: map[string]data.Value{
				"collection":   {Kind: data.KindString, S: "c"},
				"chunks_added": {Kind: data.KindInt, I: 3},
				"total_chunks": {Kind: data.KindInt, I: 10},
				"dry_run":      {Kind: data.KindBool, B: true},
			},
		}},
	}
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	WriteRAGVerbose(nil, shape, driver.RAGTaskIndex)
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "index:") {
		t.Fatalf("stderr=%q", buf.String())
	}
}

func TestWriteRAGOutputQuietAnswer(t *testing.T) {
	shape := &data.RecordSet{
		Records: []data.Record{
			{Role: "answer", Fields: map[string]data.Value{"text": {Kind: data.KindString, S: "hello"}}},
		},
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := WriteRAGOutput(shape, data.DefaultExportOpts(), true, driver.RAGTaskAnswer)
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "hello" {
		t.Fatalf("output=%q", buf.String())
	}
}
