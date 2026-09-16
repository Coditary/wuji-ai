package clix

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/coditary/wuji-core/pkg/driver"
)

type TrainBase struct {
	Enabled    bool
	Dataset    string
	Epochs     int
	Output     string
	MethodFlag string
}

func TrainFlagSet(cmd *cobra.Command, persistent bool) *pflag.FlagSet {
	if persistent {
		return cmd.PersistentFlags()
	}
	return cmd.Flags()
}

func AddTrainBaseFlags(cmd *cobra.Command, b *TrainBase, persistent bool) {
	flags := cmd.Flags()
	if persistent {
		flags = cmd.PersistentFlags()
	}
	flags.BoolVar(&b.Enabled, "train", false, "fine-tune instead of generating (prefer: wuji train <capability> [method])")
	flags.StringVar(&b.Dataset, "dataset", "", "training dataset ID or path")
	flags.IntVar(&b.Epochs, "epochs", 10, "training epochs")
	flags.StringVar(&b.Output, "train-output", "", "output path for trained weights")
	flags.StringVar(&b.Output, "output", "", "alias for --train-output")
	flags.StringVar(&b.MethodFlag, "train-method", "", "training method (overrides inference; e.g. lora, dreambooth, speech)")
}

func RegisterTrainModelAlias(cmd *cobra.Command, baseModel *string, persistent bool) {
	if !persistent {
		return
	}
	cmd.PersistentFlags().StringVar(baseModel, "model", "", "alias for --base-model")
}

func methodFlagGuide(method string) string {
	guides := map[string]string{
		"full":              "Common: --base-model, --learning-rate, --epochs, --batch-size, --seed",
		"lora":              "Required: --lora-rank. Common: --lora-alpha, --base-model, --learning-rate",
		"qlora":             "Required: --lora-rank. Common: --lora-alpha, --base-model (4-bit quantized)",
		"dreambooth":        "Required: --class-token. Optional: --prior-preservation, --reg-dataset",
		"textual-inversion": "Required: --token. Common: --train-width, --train-height",
		"controlnet":        "Required: --control-type or --pose/--depth/…. Common: --base-model",
		"embedding":         "Common: --base-model, --train-width, --train-height",
		"music":             "Common: --base-model, --sample-rate, --batch-size",
		"melody":            "Common: --base-model, --sample-rate (melody + target pairs in dataset)",
		"sfx":               "Common: --base-model, --sample-rate",
		"speech":            "Common: --base-model, --sample-rate, --batch-size",
		"rvc":               "Common: --sample, --pretrained-model, --pitch-shift, --index-rate",
	}
	if g, ok := guides[method]; ok {
		return "\n\nFlags: " + g
	}
	return ""
}

func FormatMethodCatalog(catalog []driver.TrainMethodInfo) string {
	var b strings.Builder
	for i, info := range catalog {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("  ")
		b.WriteString(info.Method)
		b.WriteString(" — ")
		b.WriteString(info.Description)
		b.WriteString("\n    Dataset: ")
		b.WriteString(info.Dataset)
		b.WriteString(methodFlagGuide(info.Method))
	}
	return b.String()
}

func NewTrainMethodCmd(use, short string, info driver.TrainMethodInfo, run func(*cobra.Command, []string) error) *cobra.Command {
	long := info.Description
	if info.Dataset != "" {
		long += "\n\nDataset: " + info.Dataset
	}
	long += methodFlagGuide(info.Method)
	return &cobra.Command{
		Use:   use + " [name]",
		Short: short,
		Long:  long,
		Args:  cobra.ArbitraryArgs,
		RunE:  run,
	}
}
