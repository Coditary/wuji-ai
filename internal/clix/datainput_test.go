package clix

import (
	"os"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func TestParseExportFormatDefaultJSON(t *testing.T) {
	f, err := ParseExportFormat(false, false, false, false, "")
	if err != nil || f != "json" {
		t.Fatalf("format=%q err=%v", f, err)
	}
}

func TestParseExportFormatConflicting(t *testing.T) {
	_, err := ParseExportFormat(true, true, false, false, "")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestParseExportFormatInvalid(t *testing.T) {
	_, err := ParseExportFormat(false, false, false, false, "xml")
	if err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestParseExportFormatExplicit(t *testing.T) {
	f, err := ParseExportFormat(false, false, true, false, "")
	if err != nil || f != data.FormatNDJSON {
		t.Fatalf("format=%q err=%v", f, err)
	}
}

func TestDetectInputFormat(t *testing.T) {
	if DetectInputFormat("data.csv", driver.DataInputAuto) != driver.DataInputCSV {
		t.Fatal("expected csv")
	}
	if DetectInputFormat("data.mpack", driver.DataInputAuto) != driver.DataInputMsgpack {
		t.Fatal("expected msgpack")
	}
	if DetectInputFormat("data.json", driver.DataInputAuto) != driver.DataInputJSON {
		t.Fatal("expected json")
	}
	if DetectInputFormat("x.bin", driver.DataInputCSV) != driver.DataInputCSV {
		t.Fatal("explicit format should win")
	}
}

func TestDataInputOr(t *testing.T) {
	if DataInputOr(driver.DataInputAuto, driver.DataInputJSON) != driver.DataInputJSON {
		t.Fatal("auto should use fallback")
	}
	if DataInputOr(driver.DataInputCSV, driver.DataInputJSON) != driver.DataInputCSV {
		t.Fatal("explicit should be kept")
	}
}

func TestLoadDataFromReaderJSON(t *testing.T) {
	shape, err := LoadDataFromReader(strings.NewReader(`{"records":[]}`), driver.DataInputJSON)
	if err != nil {
		t.Fatal(err)
	}
	if shape.Kind() != data.ShapeRecordSet {
		t.Fatalf("kind=%s", shape.Kind())
	}
}

func TestLoadDataFromReaderCSV(t *testing.T) {
	shape, err := LoadDataFromReader(strings.NewReader("a,b\n1,2\n"), driver.DataInputCSV)
	if err != nil {
		t.Fatal(err)
	}
	if shape.Kind() != data.ShapeTable {
		t.Fatalf("kind=%s", shape.Kind())
	}
}

func TestLoadDataFromPath(t *testing.T) {
	f, err := os.CreateTemp("", "clix-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	shape, err := LoadDataFromPath(f.Name(), driver.DataInputCSV)
	if err != nil {
		t.Fatal(err)
	}
	if shape.Kind() != data.ShapeTable {
		t.Fatalf("kind=%s", shape.Kind())
	}
}

func TestBuildDataRequestTextFile(t *testing.T) {
	f, err := os.CreateTemp("", "clix-text-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("from file"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := BuildDataRequest(nil, DataRequestFields{
		TaskFlags: DataTaskFlags{Embed: true},
		TextFile:  f.Name(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Text != "from file" {
		t.Fatalf("text=%q", req.Text)
	}
}

func TestJoinArgsUtil(t *testing.T) {
	if JoinArgs([]string{"a", "b"}) != "a b" {
		t.Fatal("join failed")
	}
}
