package clix

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func LoadDataInput(fields DataInputFields) (data.DataShape, error) {
	if fields.InputPath != "" {
		return LoadDataFromPath(fields.InputPath, fields.InputFormat)
	}
	if fields.UseStdin || StdinPiped() {
		return LoadDataFromReader(os.Stdin, fields.InputFormat)
	}
	if fields.CSVPath != "" {
		t, err := data.ParseCSVFile(fields.CSVPath)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	if fields.GraphPath != "" {
		return LoadDataFromPath(fields.GraphPath, DataInputOr(fields.InputFormat, driver.DataInputJSON))
	}
	return nil, nil
}

func LoadDataFromPath(path string, format driver.DataInputFormat) (data.DataShape, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadDataFromReader(f, DetectInputFormat(path, format))
}

func LoadDataFromReader(r io.Reader, format driver.DataInputFormat) (data.DataShape, error) {
	switch format {
	case driver.DataInputCSV:
		return data.ParseCSV(r)
	case driver.DataInputMsgpack:
		return data.ParseMsgpack(r)
	default:
		return data.ParseJSON(r)
	}
}

func DetectInputFormat(path string, explicit driver.DataInputFormat) driver.DataInputFormat {
	if explicit != "" && explicit != driver.DataInputAuto {
		return explicit
	}
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".csv"):
		return driver.DataInputCSV
	case strings.HasSuffix(lower, ".msgpack"), strings.HasSuffix(lower, ".mpack"):
		return driver.DataInputMsgpack
	default:
		return driver.DataInputJSON
	}
}

func DataInputOr(f driver.DataInputFormat, fallback driver.DataInputFormat) driver.DataInputFormat {
	if f == "" || f == driver.DataInputAuto {
		return fallback
	}
	return f
}

func ParseExportFormat(jsonOut, csvOut, ndjsonOut, msgpackOut bool, format string) (data.ExportFormat, error) {
	set := 0
	var out data.ExportFormat
	if jsonOut {
		out = data.FormatJSON
		set++
	}
	if csvOut {
		out = data.FormatCSV
		set++
	}
	if ndjsonOut {
		out = data.FormatNDJSON
		set++
	}
	if msgpackOut {
		out = data.FormatMsgpack
		set++
	}
	if format != "" {
		switch data.ExportFormat(format) {
		case data.FormatJSON, data.FormatCSV, data.FormatNDJSON, data.FormatMsgpack:
			out = data.ExportFormat(format)
			set++
		default:
			return "", errInvalidFormat(format)
		}
	}
	if set == 0 {
		return data.FormatJSON, nil
	}
	if set > 1 {
		return "", errConflictingFormats()
	}
	return out, nil
}

func errInvalidFormat(format string) error {
	return fmt.Errorf("invalid --format %q (use json, csv, ndjson, or msgpack)", format)
}

func errConflictingFormats() error {
	return fmt.Errorf("choose one output format: --json, --csv, --ndjson, --msgpack, or --format")
}

type DataInputFields struct {
	TextPath    string
	TextFile    string
	ImagePath   string
	CSVPath     string
	GraphPath   string
	InputPath   string
	UseStdin    bool
	InputFormat driver.DataInputFormat
}

func BuildDataRequest(args []string, fields DataRequestFields) (driver.DataRequest, error) {
	task, err := fields.TaskFlags.Resolve()
	if err != nil {
		return driver.DataRequest{}, err
	}

	text := strings.TrimSpace(strings.Join(args, " "))
	if fields.TextFile != "" {
		b, err := os.ReadFile(fields.TextFile)
		if err != nil {
			return driver.DataRequest{}, err
		}
		text = string(b)
	}

	inShape := fields.InputShape
	if inShape == "" {
		inShape = data.ShapeKind("auto")
	}
	outShape := fields.OutputShape
	if outShape == "" {
		outShape = data.ShapeKind("auto")
	}

	input, err := LoadDataInput(DataInputFields{
		TextFile: fields.TextFile, ImagePath: fields.ImagePath,
		CSVPath: fields.CSVPath, GraphPath: fields.GraphPath,
		InputPath: fields.InputPath, UseStdin: fields.UseStdin,
		InputFormat: fields.InputFormat,
	})
	if err != nil {
		return driver.DataRequest{}, err
	}

	req := driver.DataRequest{
		Task: task, Model: fields.Model, Text: text,
		TextFile: fields.TextFile, ImagePath: fields.ImagePath,
		CSVPath: fields.CSVPath, GraphPath: fields.GraphPath,
		UseStdin: fields.UseStdin, InputFormat: fields.InputFormat,
		InputShape: inShape, OutputShape: outShape,
		Input: input, Horizon: fields.Horizon,
	}
	return req, nil
}

type DataRequestFields struct {
	TaskFlags   DataTaskFlags
	Model       string
	TextFile    string
	ImagePath   string
	CSVPath     string
	GraphPath   string
	InputPath   string
	UseStdin    bool
	InputFormat driver.DataInputFormat
	InputShape  data.ShapeKind
	OutputShape data.ShapeKind
	Horizon     int
}

func StdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
