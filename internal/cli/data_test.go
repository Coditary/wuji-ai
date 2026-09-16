package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func TestParseExportFormatDefaultJSON(t *testing.T) {
	f, err := parseExportFormat(false, false, false, false, "")
	if err != nil || f != "json" {
		t.Fatalf("format=%q err=%v", f, err)
	}
}

func TestDataEmbedDummy(t *testing.T) {
	app := newTestApp(t)
	req, err := buildDataRequest([]string{"Hallo Welt"}, dataRequestFields{
		taskFlags: dataTaskFlags{embed: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	shape, err := app.Core.ProduceData(context.Background(), app.Core.DefaultDriverID(), req)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := shape.Export(&buf, data.DefaultExportOpts()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "embedding") {
		t.Fatalf("output: %s", buf.String())
	}
}

func TestDataConvertCSVToJSON(t *testing.T) {
	app := newTestApp(t)
	path := writeTempCSV(t, "a,b\n1,2\n")
	defer os.Remove(path)

	table, err := loadDataFromPath(path, driver.DataInputCSV)
	if err != nil {
		t.Fatal(err)
	}
	req := driver.DataRequest{
		InputFormat: driver.DataInputCSV,
		OutputShape: "table",
		Input:       table,
	}

	shape, err := app.Core.ProduceData(context.Background(), "", req)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := shape.Export(&buf, data.DefaultExportOpts()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "row0") {
		t.Fatalf("output: %s", buf.String())
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "wuji-data-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}
