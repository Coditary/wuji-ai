package clix

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/spf13/cobra"
)

func TestAddDataTaskFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "data"}
	var f DataTaskFlags
	AddDataTaskFlags(cmd, &f)
	if err := cmd.ParseFlags([]string{"--embed"}); err != nil {
		t.Fatal(err)
	}
	task, err := f.Resolve()
	if err != nil || task != driver.DataTaskEmbed {
		t.Fatalf("task=%s err=%v", task, err)
	}
}

func TestAddRAGOpFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "rag"}
	var f RAGOpFlags
	AddRAGOpFlags(cmd, &f)
	if err := cmd.ParseFlags([]string{"--query"}); err != nil {
		t.Fatal(err)
	}
	task, err := f.Resolve()
	if err != nil || task != driver.RAGTaskQuery {
		t.Fatalf("task=%s err=%v", task, err)
	}
}

func TestTrainFlagHelpers(t *testing.T) {
	cmd := &cobra.Command{Use: "train"}
	var base TrainBase
	AddTrainBaseFlags(cmd, &base, false)
	if TrainFlagSet(cmd, false) == nil {
		t.Fatal("expected flag set")
	}
	if err := cmd.ParseFlags([]string{"--train", "--dataset", "d", "--epochs", "3"}); err != nil {
		t.Fatal(err)
	}
	if !base.Enabled || base.Dataset != "d" || base.Epochs != 3 {
		t.Fatalf("base=%+v", base)
	}

	var model string
	RegisterTrainModelAlias(cmd, &model, true)
	if err := cmd.ParseFlags([]string{"--model", "m"}); err != nil {
		t.Fatal(err)
	}
	if model != "m" {
		t.Fatalf("model=%q", model)
	}

	info := driver.TrainMethodInfo{Method: "full", Description: "full fine-tune", Dataset: "pairs"}
	sub := NewTrainMethodCmd("full", "full", info, func(*cobra.Command, []string) error { return nil })
	if sub.Long == "" || sub.Use != "full [name]" {
		t.Fatalf("cmd=%+v", sub)
	}
}

func TestAddAudioInferenceFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "text"}
	var f AudioInferenceFields
	AddAudioInferenceFlags(cmd, &f)
	if err := cmd.ParseFlags([]string{"--quality", "fast", "--trim-silence", "--silence-level", "0.5"}); err != nil {
		t.Fatal(err)
	}
	if f.Quality != "fast" || !f.TrimSilence || f.SilenceLevel != 0.5 {
		t.Fatalf("fields=%+v", f)
	}
}
