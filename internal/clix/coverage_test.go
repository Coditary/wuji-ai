package clix

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/spf13/cobra"
)

func TestLoadDataInputCSVPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.csv")
	if err := os.WriteFile(path, []byte("a,b\n1,2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	shape, err := LoadDataInput(DataInputFields{CSVPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if shape.Kind() != data.ShapeTable {
		t.Fatalf("kind=%s", shape.Kind())
	}
}

func TestLoadDataInputGraphPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "g.json")
	if err := os.WriteFile(path, []byte(`{"records":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	shape, err := LoadDataInput(DataInputFields{GraphPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if shape == nil {
		t.Fatal("expected shape")
	}
}

func TestBuildImageRequestTrainAndSprite(t *testing.T) {
	req, err := BuildImageRequest("p", true, ImageRequestFields{Model: "m"})
	if err != nil || req.Model != "m" {
		t.Fatalf("req=%+v err=%v", req, err)
	}

	req, err = BuildImageRequest("walk", false, ImageRequestFields{
		SpriteRequested: true,
		BatchTotal:      8,
		Width:           64,
		Height:          64,
		AvailableVRAMMB: 8000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Task != driver.ImageTaskSprite || req.FrameCount <= 0 {
		t.Fatalf("sprite req=%+v", req)
	}

	req, err = BuildImageRequest("x", false, ImageRequestFields{
		ControlType: "pose",
		ControlUnits: []driver.ControlNetUnit{{
			Type: driver.ImageControlDepth, ImagePath: "d.png",
		}},
		Seed: 42,
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.ControlType != driver.ImageControlPose || req.Seed == nil {
		t.Fatalf("req=%+v", req)
	}
}

func TestBuildVideoAndMeshTrainModes(t *testing.T) {
	vreq, err := BuildVideoRequest("p", true, VideoRequestFields{Model: "vm"})
	if err != nil || vreq.Model != "vm" {
		t.Fatalf("vreq=%+v err=%v", vreq, err)
	}
	mreq, err := BuildMeshRequest("p", true, MeshRequestFields{Format: "glb"})
	if err != nil || mreq.Format != "glb" {
		t.Fatalf("mreq=%+v err=%v", mreq, err)
	}
}

func TestResolveMethodsFromMethodFlag(t *testing.T) {
	if ResolveTextMethod(TextTrainOpts{MethodFlag: "qlora"}) != driver.TextTrainMethodQLoRA {
		t.Fatal("method flag qlora")
	}
	if ResolveImageMethod(ImageTrainOpts{Token: "tok"}) != driver.ImageTrainMethodTextualInversion {
		t.Fatal("textual inversion")
	}
	if ResolveImageMethod(ImageTrainOpts{ControlType: "pose"}) != driver.ImageTrainMethodControlNet {
		t.Fatal("controlnet")
	}
	if ResolveVideoMethod(VideoTrainOpts{MethodFlag: "lora"}) != driver.VideoTrainMethodLoRA {
		t.Fatal("video lora flag")
	}
	if ResolveMeshMethod(MeshTrainOpts{MethodFlag: "lora"}) != driver.MeshTrainMethodLoRA {
		t.Fatal("mesh lora flag")
	}
	if ResolveVoiceMethod(VoiceTrainOpts{MethodFlag: "rvc"}) != driver.VoiceTrainMethodRVC {
		t.Fatal("voice rvc flag")
	}
}

func TestResolveRAGQueryFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "q.txt")
	if err := os.WriteFile(path, []byte("file query"), 0o644); err != nil {
		t.Fatal(err)
	}
	q, err := resolveRAGQuery(RAGRequestFields{QueryFile: path})
	if err != nil || q != "file query" {
		t.Fatalf("q=%q err=%v", q, err)
	}
}

func TestWriteRAGVerboseStatsAndQuery(t *testing.T) {
	shape := &data.RecordSet{
		Records: []data.Record{
			{Role: "chunk"}, {Role: "chunk"},
			{Role: "collection", Fields: map[string]data.Value{
				"collection":  {Kind: data.KindString, S: "c"},
				"dims":        {Kind: data.KindInt, I: 768},
				"chunk_count": {Kind: data.KindInt, I: 2},
				"embed_model": {Kind: data.KindString, S: "ollama"},
				"updated_at":  {Kind: data.KindString, S: "now"},
			}},
		},
	}
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	WriteRAGVerbose(nil, shape, driver.RAGTaskStats)
	WriteRAGVerbose(nil, shape, driver.RAGTaskQuery)
	w.Close()
	os.Stderr = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	outStr := string(out)
	if !strings.Contains(outStr, "collection:") || !strings.Contains(outStr, "retrieval:") {
		t.Fatalf("stderr=%q", outStr)
	}
}

func TestControlModeInvalid(t *testing.T) {
	cmd := &cobra.Command{Use: "image"}
	opts := &ImageControlOptions{}
	RegisterImageControlFlags(cmd, opts)
	if err := cmd.ParseFlags([]string{"--control-mode", "invalid"}); err != nil {
		t.Fatal(err)
	}
	_, err := opts.ControlMode()
	if err == nil {
		t.Fatal("expected invalid control mode error")
	}
}

func TestReadPasswordNonTTY(t *testing.T) {
	old := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.WriteString("secret\n")
		w.Close()
	}()
	pw, err := ReadPassword("pw: ")
	os.Stdin = old
	if err != nil || pw != "secret" {
		t.Fatalf("pw=%q err=%v", pw, err)
	}
}

func TestFirstControlTypeEmpty(t *testing.T) {
	if FirstControlType(nil) != "" {
		t.Fatal("expected empty")
	}
}

func TestParseImageScaleInputError(t *testing.T) {
	_, _, err := ParseImageScaleInput("not-a-scale", false)
	if err == nil {
		t.Fatal("expected parse error")
	}
}
