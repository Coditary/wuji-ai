package clix

import (
	"fmt"
	"os"
	"strings"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func ApplyRAGOutputShape(shape data.DataShape, outputShape string) data.DataShape {
	if outputShape != string(driver.RAGOutputAnswerOnly) {
		return shape
	}
	rs, ok := shape.(*data.RecordSet)
	if !ok || rs == nil {
		return shape
	}
	out := &data.RecordSet{MetaData: rs.MetaData}
	for _, rec := range rs.Records {
		if rec.Role == "answer" {
			out.Records = append(out.Records, rec)
		}
	}
	return out
}

func WriteRAGOutput(shape data.DataShape, opts data.ExportOpts, quiet bool, task driver.RAGTask) error {
	if quiet && task == driver.RAGTaskAnswer {
		rs, ok := shape.(*data.RecordSet)
		if !ok {
			return shape.Export(os.Stdout, opts)
		}
		for _, rec := range rs.Records {
			if rec.Role == "answer" {
				if v, ok := rec.Fields["text"]; ok {
					_, err := fmt.Fprintln(os.Stdout, v.S)
					return err
				}
			}
		}
	}
	return shape.Export(os.Stdout, opts)
}

func WriteRAGVerbose(stderr *os.File, shape data.DataShape, task driver.RAGTask) {
	if stderr == nil {
		stderr = os.Stderr
	}
	rs, ok := shape.(*data.RecordSet)
	if !ok {
		return
	}
	switch task {
	case driver.RAGTaskIndex:
		for _, rec := range rs.Records {
			if rec.Role != "result" {
				continue
			}
			fmt.Fprintf(stderr, "index: collection=%s chunks_added=%v total=%v dry_run=%v\n",
				FieldString(rec, "collection"), FieldString(rec, "chunks_added"),
				FieldString(rec, "total_chunks"), FieldString(rec, "dry_run"))
		}
	case driver.RAGTaskQuery, driver.RAGTaskAnswer:
		count := 0
		for _, rec := range rs.Records {
			if rec.Role == "chunk" {
				count++
			}
		}
		fmt.Fprintf(stderr, "retrieval: %d chunks\n", count)
	case driver.RAGTaskStats, driver.RAGTaskInfo:
		for _, rec := range rs.Records {
			if rec.Role != "collection" {
				continue
			}
			fmt.Fprintf(stderr, "collection: name=%s dims=%s chunks=%s embed_model=%s updated=%s\n",
				FieldString(rec, "collection"), FieldString(rec, "dims"),
				FieldString(rec, "chunk_count"), FieldString(rec, "embed_model"),
				FieldString(rec, "updated_at"))
		}
	}
}

func FieldString(rec data.Record, name string) string {
	if v, ok := rec.Fields[name]; ok {
		if v.Kind == data.KindString {
			return v.S
		}
		if v.Kind == data.KindInt {
			return fmt.Sprintf("%d", v.I)
		}
		if v.Kind == data.KindBool {
			return fmt.Sprintf("%t", v.B)
		}
	}
	return ""
}

func ParseCSVFields(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
