package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
)

func loadDataInput(fields dataInputFields) (data.DataShape, error) {
	if fields.inputPath != "" {
		return loadDataFromPath(fields.inputPath, fields.inputFormat)
	}
	if fields.useStdin || stdinPiped() {
		return loadDataFromReader(os.Stdin, fields.inputFormat)
	}
	if fields.csvPath != "" {
		t, err := data.ParseCSVFile(fields.csvPath)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	if fields.graphPath != "" {
		return loadDataFromPath(fields.graphPath, dataInputOr(fields.inputFormat, driver.DataInputJSON))
	}
	return nil, nil
}

func loadDataFromPath(path string, format driver.DataInputFormat) (data.DataShape, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadDataFromReader(f, detectInputFormat(path, format))
}

func loadDataFromReader(r io.Reader, format driver.DataInputFormat) (data.DataShape, error) {
	switch format {
	case driver.DataInputCSV:
		return data.ParseCSV(r)
	case driver.DataInputMsgpack:
		return data.ParseMsgpack(r)
	default:
		return data.ParseJSON(r)
	}
}

func detectInputFormat(path string, explicit driver.DataInputFormat) driver.DataInputFormat {
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

func dataInputOr(f driver.DataInputFormat, fallback driver.DataInputFormat) driver.DataInputFormat {
	if f == "" || f == driver.DataInputAuto {
		return fallback
	}
	return f
}

func parseExportFormat(jsonOut, csvOut, ndjsonOut, msgpackOut bool, format string) (data.ExportFormat, error) {
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

type dataInputFields struct {
	textPath   string
	textFile   string
	imagePath  string
	csvPath    string
	graphPath  string
	inputPath  string
	useStdin   bool
	inputFormat driver.DataInputFormat
}

func buildDataRequest(args []string, fields dataRequestFields) (driver.DataRequest, error) {
	task, err := fields.taskFlags.resolve()
	if err != nil {
		return driver.DataRequest{}, err
	}

	text := strings.TrimSpace(strings.Join(args, " "))
	if fields.textFile != "" {
		b, err := os.ReadFile(fields.textFile)
		if err != nil {
			return driver.DataRequest{}, err
		}
		text = string(b)
	}

	inShape := fields.inputShape
	if inShape == "" {
		inShape = data.ShapeKind("auto")
	}
	outShape := fields.outputShape
	if outShape == "" {
		outShape = data.ShapeKind("auto")
	}

	input, err := loadDataInput(dataInputFields{
		textFile: fields.textFile, imagePath: fields.imagePath,
		csvPath: fields.csvPath, graphPath: fields.graphPath,
		inputPath: fields.inputPath, useStdin: fields.useStdin,
		inputFormat: fields.inputFormat,
	})
	if err != nil {
		return driver.DataRequest{}, err
	}

	req := driver.DataRequest{
		Task: task, Model: fields.model, Text: text,
		TextFile: fields.textFile, ImagePath: fields.imagePath,
		CSVPath: fields.csvPath, GraphPath: fields.graphPath,
		UseStdin: fields.useStdin, InputFormat: fields.inputFormat,
		InputShape: inShape, OutputShape: outShape,
		Input: input, Horizon: fields.horizon,
	}
	return req, nil
}

type dataRequestFields struct {
	taskFlags    dataTaskFlags
	model        string
	textFile     string
	imagePath    string
	csvPath      string
	graphPath    string
	inputPath    string
	useStdin     bool
	inputFormat  driver.DataInputFormat
	inputShape   data.ShapeKind
	outputShape  data.ShapeKind
	horizon      int
}

func stdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
