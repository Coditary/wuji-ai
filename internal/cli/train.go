package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-ai/internal/clix"
	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/driver"
)

type trainBase = clix.TrainBase

func requirePromptUnlessTrain(train bool, args []string) error {
	if train || len(args) >= 1 {
		return nil
	}
	return fmt.Errorf("name required (positional argument)")
}

func printTrainResp(resp *driver.TrainResponse) {
	fmt.Fprintf(os.Stdout, "training job %s [%s]: %s\n", resp.JobID, resp.Capability, resp.Status)
	if resp.OutputPath != "" {
		fmt.Fprintf(os.Stdout, "output: %s\n", resp.OutputPath)
	}
}

// --- Text ---

type textTrainOpts struct {
	trainBase
	Method                    driver.TextTrainMethod
	BaseModel                 string
	LearningRate              float32
	LoRARank                  int
	LoRAAlpha                 int
	ContextLength             int
	BatchSize                 int
	GradientAccumulationSteps int
	WarmupSteps               int
	SaveEveryEpoch            int
	Seed                      int
	QLoRA                     bool
}

func addTextTrainFlags(cmd *cobra.Command, o *textTrainOpts, persistent bool) {
	clix.AddTrainBaseFlags(cmd, &o.trainBase, persistent)
	flags := clix.TrainFlagSet(cmd, persistent)
	flags.StringVar(&o.BaseModel, "base-model", "", "base checkpoint to fine-tune")
	clix.RegisterTrainModelAlias(cmd, &o.BaseModel, persistent)
	flags.Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate (0 = backend default)")
	flags.IntVar(&o.LoRARank, "lora-rank", 0, "LoRA rank (required for lora/qlora)")
	flags.IntVar(&o.LoRAAlpha, "lora-alpha", 0, "LoRA alpha / network alpha (0 = backend default)")
	flags.IntVar(&o.ContextLength, "context-length", 0, "training context length")
	flags.IntVar(&o.GradientAccumulationSteps, "gradient-accumulation", 0, "gradient accumulation steps")
	flags.IntVar(&o.WarmupSteps, "warmup-steps", 0, "learning-rate warmup steps")
	flags.IntVar(&o.SaveEveryEpoch, "save-every-epoch", 0, "checkpoint every N epochs (0 = backend default)")
	flags.BoolVar(&o.QLoRA, "qlora", false, "use QLoRA (4-bit base + adapter); sets method to qlora")
	if persistent {
		flags.IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
		flags.IntVar(&o.Seed, "seed", -1, "random seed (-1 = random)")
	}
}

func runTextTrain(ctx context.Context, app *App, driverID string, o textTrainOpts, name, defaultBase string) error {
	method := clix.ResolveTextMethod(clix.TextTrainOpts{
		Method: o.Method, MethodFlag: o.MethodFlag, QLoRA: o.QLoRA, LoRARank: o.LoRARank,
	})
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	req := driver.TextTrainRequest{
		Method: method, Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, LoRARank: o.LoRARank, LoRAAlpha: o.LoRAAlpha,
		ContextLength: o.ContextLength, BatchSize: o.BatchSize,
		GradientAccumulationSteps: o.GradientAccumulationSteps, WarmupSteps: o.WarmupSteps,
		SaveEveryEpoch: o.SaveEveryEpoch, Seed: o.Seed,
	}
	if err := driver.ValidateTextTrainRequest(req); err != nil {
		return err
	}
	resp, err := app.Core.TrainText(ctx, driverID, req)
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

// --- Image ---

type imageTrainOpts struct {
	trainBase
	Method            driver.ImageTrainMethod
	BaseModel         string
	LearningRate      float32
	LoRARank          int
	LoRAAlpha         int
	Width             int
	Height            int
	ClassToken        string
	ControlType       string
	Token             string
	Mode              string
	PriorPreservation bool
	RegDataset        string
	BatchSize         int
	Seed              int
	controls          clix.ImageControlOptions
}

func addImageTrainFlags(cmd *cobra.Command, o *imageTrainOpts, persistent bool) {
	clix.AddTrainBaseFlags(cmd, &o.trainBase, persistent)
	flags := clix.TrainFlagSet(cmd, persistent)
	flags.StringVar(&o.BaseModel, "base-model", "", "base diffusion checkpoint")
	clix.RegisterTrainModelAlias(cmd, &o.BaseModel, persistent)
	flags.Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	flags.IntVar(&o.LoRARank, "lora-rank", 0, "LoRA rank (required for lora)")
	flags.IntVar(&o.LoRAAlpha, "lora-alpha", 0, "LoRA alpha (0 = backend default)")
	flags.StringVar(&o.ClassToken, "class-token", "", "DreamBooth class token (required for dreambooth)")
	flags.StringVar(&o.Token, "token", "", "trigger token (required for textual-inversion)")
	flags.BoolVar(&o.PriorPreservation, "prior-preservation", false, "DreamBooth prior-preservation loss")
	flags.StringVar(&o.RegDataset, "reg-dataset", "", "regularization/class images dataset for DreamBooth")
	if persistent {
		flags.StringVar(&o.ControlType, "control-type", "", "ControlNet type (pose, depth, canny, …)")
		flags.IntVarP(&o.Width, "train-width", "W", 0, "training resolution width")
		flags.IntVarP(&o.Height, "train-height", "H", 0, "training resolution height")
		flags.StringVar(&o.Mode, "mode", "", "pipeline profile: prop, pixel, tile, ui, character, texture, …")
		flags.IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
		flags.IntVar(&o.Seed, "seed", -1, "random seed (-1 = random)")
		clix.RegisterImageControlFlags(cmd, &o.controls)
	} else {
		flags.IntVar(&o.Width, "train-width", 0, "training resolution width (defaults to -W)")
		flags.IntVar(&o.Height, "train-height", 0, "training resolution height (defaults to -H)")
	}
}

func runImageTrain(ctx context.Context, app *App, driverID string, o imageTrainOpts, name, defaultBase string, cmd *cobra.Command) error {
	method := clix.ResolveImageMethod(clix.ImageTrainOpts{
		Method: o.Method, MethodFlag: o.MethodFlag, ClassToken: o.ClassToken,
		Token: o.Token, ControlType: o.ControlType, LoRARank: o.LoRARank,
	})
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}

	ct := o.ControlType
	if ct == "" && cmd != nil {
		units, err := o.controls.BuildUnits(cmd)
		if err != nil {
			return err
		}
		if t := clix.FirstControlType(units); t != "" {
			ct = string(t)
		}
	}

	var controlType driver.ImageControlType
	var mode driver.ImageMode
	var err error
	if ct != "" {
		controlType, err = driver.ParseImageControlType(ct)
		if err != nil {
			return err
		}
	}
	if o.Mode != "" {
		mode, err = driver.ParseImageMode(o.Mode)
		if err != nil {
			return err
		}
	}

	req := driver.ImageTrainRequest{
		Method: method, Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, LoRARank: o.LoRARank, LoRAAlpha: o.LoRAAlpha,
		Width: o.Width, Height: o.Height, ClassToken: o.ClassToken,
		ControlType: controlType, Token: o.Token, Mode: mode,
		PriorPreservation: o.PriorPreservation, RegDatasetID: o.RegDataset,
		BatchSize: o.BatchSize, Seed: o.Seed,
	}
	if err := driver.ValidateImageTrainRequest(req); err != nil {
		return err
	}
	resp, err := app.Core.TrainImage(ctx, driverID, req)
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

// --- Video ---

type videoTrainOpts struct {
	trainBase
	Method        driver.VideoTrainMethod
	BaseModel     string
	LearningRate  float32
	LoRARank      int
	Frames        int
	FPS           int
	ContextLength int
	BatchSize     int
	Seed          int
}

func addVideoTrainFlags(cmd *cobra.Command, o *videoTrainOpts, persistent bool) {
	clix.AddTrainBaseFlags(cmd, &o.trainBase, persistent)
	flags := clix.TrainFlagSet(cmd, persistent)
	flags.StringVar(&o.BaseModel, "base-model", "", "base video model checkpoint")
	clix.RegisterTrainModelAlias(cmd, &o.BaseModel, persistent)
	flags.Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	flags.IntVar(&o.LoRARank, "lora-rank", 0, "LoRA rank (required for lora)")
	flags.IntVar(&o.Frames, "train-frames", 0, "training clip length in frames")
	flags.IntVar(&o.FPS, "train-fps", 0, "training frame rate")
	flags.IntVar(&o.ContextLength, "train-context-length", 0, "temporal context window for training")
	if persistent {
		flags.IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
		flags.IntVar(&o.Seed, "seed", -1, "random seed (-1 = random)")
	}
}

func runVideoTrain(ctx context.Context, app *App, driverID string, o videoTrainOpts, name, defaultBase string) error {
	method := clix.ResolveVideoMethod(clix.VideoTrainOpts{
		Method: o.Method, MethodFlag: o.MethodFlag, LoRARank: o.LoRARank,
	})
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	req := driver.VideoTrainRequest{
		Method: method, Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, LoRARank: o.LoRARank,
		Frames: o.Frames, FPS: o.FPS, ContextLength: o.ContextLength,
		BatchSize: o.BatchSize, Seed: o.Seed,
	}
	if err := driver.ValidateVideoTrainRequest(req); err != nil {
		return err
	}
	resp, err := app.Core.TrainVideo(ctx, driverID, req)
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

// --- Audio ---

type audioTrainOpts struct {
	trainBase
	Method       driver.AudioTrainMethod
	BaseModel    string
	LearningRate float32
	SampleRate   int
	BatchSize    int
	Seed         int
}

func addAudioTrainFlags(cmd *cobra.Command, o *audioTrainOpts, persistent bool) {
	clix.AddTrainBaseFlags(cmd, &o.trainBase, persistent)
	flags := clix.TrainFlagSet(cmd, persistent)
	flags.StringVar(&o.BaseModel, "base-model", "", "base audio model checkpoint")
	clix.RegisterTrainModelAlias(cmd, &o.BaseModel, persistent)
	flags.Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	if persistent {
		flags.IntVar(&o.SampleRate, "sample-rate", 0, "target audio sample rate (Hz)")
		flags.IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
		flags.IntVar(&o.Seed, "seed", -1, "random seed (-1 = random)")
	}
}

func runAudioTrain(ctx context.Context, app *App, driverID string, o audioTrainOpts, name, defaultBase string, sfx, speech bool, referencePath string) error {
	method := clix.ResolveAudioMethod(clix.AudioTrainOpts{
		Method: o.Method, MethodFlag: o.MethodFlag,
	}, sfx, speech, referencePath)
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	req := driver.AudioTrainRequest{
		Method: method, Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, SampleRate: o.SampleRate,
		BatchSize: o.BatchSize, TaskType: method.ToAudioTask(), Seed: o.Seed,
	}
	if err := driver.ValidateAudioTrainRequest(req); err != nil {
		return err
	}
	resp, err := app.Core.TrainAudio(ctx, driverID, req)
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

// --- Mesh ---

type meshTrainOpts struct {
	trainBase
	Method       driver.MeshTrainMethod
	BaseModel    string
	LearningRate float32
	LoRARank     int
	Format       string
	Mode         string
	Seed         int
}

func addMeshTrainFlags(cmd *cobra.Command, o *meshTrainOpts, persistent bool) {
	clix.AddTrainBaseFlags(cmd, &o.trainBase, persistent)
	flags := clix.TrainFlagSet(cmd, persistent)
	flags.StringVar(&o.BaseModel, "base-model", "", "base mesh model checkpoint")
	clix.RegisterTrainModelAlias(cmd, &o.BaseModel, persistent)
	flags.Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	flags.IntVar(&o.LoRARank, "lora-rank", 0, "LoRA rank (required for lora)")
	flags.StringVar(&o.Format, "train-format", "", "output format for trained weights")
	if persistent {
		flags.StringVar(&o.Mode, "mode", "", "pipeline profile: prop, human, hardsurface, terrain, …")
		flags.IntVar(&o.Seed, "seed", -1, "random seed (-1 = random)")
	}
}

func runMeshTrain(ctx context.Context, app *App, driverID string, o meshTrainOpts, name, defaultBase string) error {
	method := clix.ResolveMeshMethod(clix.MeshTrainOpts{
		Method: o.Method, MethodFlag: o.MethodFlag, LoRARank: o.LoRARank,
	})
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	var mode driver.MeshMode
	if o.Mode != "" {
		var err error
		mode, err = driver.ParseMeshMode(o.Mode)
		if err != nil {
			return err
		}
	}
	req := driver.MeshTrainRequest{
		Method: method, Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, LoRARank: o.LoRARank, Format: o.Format,
		Mode: mode, Seed: o.Seed,
	}
	if err := driver.ValidateMeshTrainRequest(req); err != nil {
		return err
	}
	resp, err := app.Core.TrainMesh(ctx, driverID, req)
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

// --- Voice ---

type voiceTrainOpts struct {
	trainBase
	Method          driver.VoiceTrainMethod
	SamplePath      string
	PretrainedModel string
	PitchShift      int
	IndexRate       float32
	BatchSize       int
	SaveEveryEpoch  int
	Seed            int
}

func addVoiceTrainFlags(cmd *cobra.Command, o *voiceTrainOpts, persistent bool) {
	clix.AddTrainBaseFlags(cmd, &o.trainBase, persistent)
	flags := clix.TrainFlagSet(cmd, persistent)
	flags.StringVar(&o.SamplePath, "sample", "", "reference voice sample (WAV)")
	flags.StringVar(&o.PretrainedModel, "pretrained-model", "", "pretrained RVC base model")
	flags.IntVar(&o.PitchShift, "pitch-shift", 0, "pitch shift in semitones")
	flags.Float32Var(&o.IndexRate, "index-rate", 0, "feature retrieval strength (0 = backend default)")
	flags.IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
	flags.IntVar(&o.SaveEveryEpoch, "save-every-epoch", 0, "checkpoint every N epochs (0 = backend default)")
	flags.IntVar(&o.Seed, "seed", -1, "random seed (-1 = random)")
}

func runVoiceTrain(ctx context.Context, app *App, driverID string, o voiceTrainOpts, name string) error {
	method := clix.ResolveVoiceMethod(clix.VoiceTrainOpts{
		Method: o.Method, MethodFlag: o.MethodFlag,
	})
	req := driver.VoiceTrainRequest{
		Method: method, Name: name, DatasetID: o.Dataset, OutputPath: o.Output,
		Epochs: o.Epochs, SamplePath: o.SamplePath, PretrainedModel: o.PretrainedModel,
		PitchShift: o.PitchShift, IndexRate: o.IndexRate, BatchSize: o.BatchSize,
		SaveEveryEpoch: o.SaveEveryEpoch, Seed: o.Seed,
	}
	if err := driver.ValidateVoiceTrainRequest(req); err != nil {
		return err
	}
	resp, err := app.Core.TrainVoice(ctx, driverID, req)
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

// --- Train command tree ---

func newTrainCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:   "train",
		Short: "Fine-tune models by capability and method",
		Long: `Train or fine-tune AI models.

  wuji train image my-style --dataset ./captions
      → full model fine-tune (default)

  wuji train image lora my-style --dataset ./data --lora-rank 32
  wuji train image dreambooth hero --dataset ./photos --class-token "a photo of sks person"
  wuji train text qlora my-chat --dataset ./jsonl --lora-rank 16
  wuji train audio speech my-voice --dataset ./wavs
  wuji train voice rvc my-speaker --dataset ./speech --sample ref.wav

Aliases: --model (= --base-model), --output (= --train-output), -W/-H (= train width/height).
Use --train-method to override method inference on capability --train shortcuts.`,
	}

	root.AddCommand(
		newTextTrainCmd(app),
		newImageTrainCmd(app),
		newVideoTrainCmd(app),
		newAudioTrainCmd(app),
		newMeshTrainCmd(app),
		newVoiceTrainCmd(app),
	)
	return root
}

func addCapabilityTrainSubcommands(cap *cobra.Command, catalog []driver.TrainMethodInfo, runDefault func(*cobra.Command, []string) error, runMethod func(driver.TrainMethodInfo) func(*cobra.Command, []string) error) {
	cap.RunE = runDefault
	_ = cap.MarkPersistentFlagRequired("dataset")
	for _, info := range catalog {
		cap.AddCommand(clix.NewTrainMethodCmd(info.Method, info.Description, info, runMethod(info)))
	}
}

func newTextTrainCmd(app *App) *cobra.Command {
	var opts textTrainOpts
	cap := &cobra.Command{
		Use:   "text [name]",
		Short: "Train text / LLM models",
		Long:  "Methods:\n" + clix.FormatMethodCatalog(driver.TextTrainMethodCatalog()),
		Args:  cobra.ArbitraryArgs,
	}
	addTextTrainFlags(cap, &opts, true)
	addCapabilityTrainSubcommands(cap, driver.TextTrainMethodCatalog(),
		func(cmd *cobra.Command, args []string) error {
			opts.Method = driver.TextTrainMethodFull
			return runTextTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.TextGeneration), opts, clix.JoinArgs(args), opts.BaseModel)
		},
		func(info driver.TrainMethodInfo) func(*cobra.Command, []string) error {
			method, _ := driver.ParseTextTrainMethod(info.Method)
			return func(cmd *cobra.Command, args []string) error {
				opts.Method = method
				return runTextTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.TextGeneration), opts, clix.JoinArgs(args), opts.BaseModel)
			}
		},
	)
	return cap
}

func newImageTrainCmd(app *App) *cobra.Command {
	var opts imageTrainOpts
	cap := &cobra.Command{
		Use:   "image [name]",
		Short: "Train image / diffusion models",
		Long:  "Methods:\n" + clix.FormatMethodCatalog(driver.ImageTrainMethodCatalog()),
		Args:  cobra.ArbitraryArgs,
	}
	addImageTrainFlags(cap, &opts, true)
	addCapabilityTrainSubcommands(cap, driver.ImageTrainMethodCatalog(),
		func(cmd *cobra.Command, args []string) error {
			opts.Method = driver.ImageTrainMethodFull
			return runImageTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.ImageGeneration), opts, clix.JoinArgs(args), opts.BaseModel, cmd)
		},
		func(info driver.TrainMethodInfo) func(*cobra.Command, []string) error {
			method, _ := driver.ParseImageTrainMethod(info.Method)
			return func(cmd *cobra.Command, args []string) error {
				opts.Method = method
				return runImageTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.ImageGeneration), opts, clix.JoinArgs(args), opts.BaseModel, cmd)
			}
		},
	)
	return cap
}

func newVideoTrainCmd(app *App) *cobra.Command {
	var opts videoTrainOpts
	cap := &cobra.Command{
		Use:   "video [name]",
		Short: "Train video models",
		Long:  "Methods:\n" + clix.FormatMethodCatalog(driver.VideoTrainMethodCatalog()),
		Args:  cobra.ArbitraryArgs,
	}
	addVideoTrainFlags(cap, &opts, true)
	addCapabilityTrainSubcommands(cap, driver.VideoTrainMethodCatalog(),
		func(cmd *cobra.Command, args []string) error {
			opts.Method = driver.VideoTrainMethodFull
			return runVideoTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.VideoGeneration), opts, clix.JoinArgs(args), opts.BaseModel)
		},
		func(info driver.TrainMethodInfo) func(*cobra.Command, []string) error {
			method, _ := driver.ParseVideoTrainMethod(info.Method)
			return func(cmd *cobra.Command, args []string) error {
				opts.Method = method
				return runVideoTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.VideoGeneration), opts, clix.JoinArgs(args), opts.BaseModel)
			}
		},
	)
	return cap
}

func newAudioTrainCmd(app *App) *cobra.Command {
	var opts audioTrainOpts
	cap := &cobra.Command{
		Use:   "audio [name]",
		Short: "Train audio models",
		Long:  "Methods:\n" + clix.FormatMethodCatalog(driver.AudioTrainMethodCatalog()),
		Args:  cobra.ArbitraryArgs,
	}
	addAudioTrainFlags(cap, &opts, true)
	addCapabilityTrainSubcommands(cap, driver.AudioTrainMethodCatalog(),
		func(cmd *cobra.Command, args []string) error {
			opts.Method = driver.AudioTrainMethodMusic
			return runAudioTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.AudioGeneration), opts, clix.JoinArgs(args), opts.BaseModel, false, false, "")
		},
		func(info driver.TrainMethodInfo) func(*cobra.Command, []string) error {
			method, _ := driver.ParseAudioTrainMethod(info.Method)
			return func(cmd *cobra.Command, args []string) error {
				opts.Method = method
				return runAudioTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.AudioGeneration), opts, clix.JoinArgs(args), opts.BaseModel, false, false, "")
			}
		},
	)
	return cap
}

func newMeshTrainCmd(app *App) *cobra.Command {
	var opts meshTrainOpts
	cap := &cobra.Command{
		Use:   "mesh [name]",
		Short: "Train mesh / 3D models",
		Long:  "Methods:\n" + clix.FormatMethodCatalog(driver.MeshTrainMethodCatalog()),
		Args:  cobra.ArbitraryArgs,
	}
	addMeshTrainFlags(cap, &opts, true)
	addCapabilityTrainSubcommands(cap, driver.MeshTrainMethodCatalog(),
		func(cmd *cobra.Command, args []string) error {
			opts.Method = driver.MeshTrainMethodFull
			return runMeshTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.Mesh), opts, clix.JoinArgs(args), opts.BaseModel)
		},
		func(info driver.TrainMethodInfo) func(*cobra.Command, []string) error {
			method, _ := driver.ParseMeshTrainMethod(info.Method)
			return func(cmd *cobra.Command, args []string) error {
				opts.Method = method
				return runMeshTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.Mesh), opts, clix.JoinArgs(args), opts.BaseModel)
			}
		},
	)
	return cap
}

func newVoiceTrainCmd(app *App) *cobra.Command {
	var opts voiceTrainOpts
	cap := &cobra.Command{
		Use:   "voice [name]",
		Short: "Train voice cloning models",
		Long:  "Methods:\n" + clix.FormatMethodCatalog(driver.VoiceTrainMethodCatalog()),
		Args:  cobra.ArbitraryArgs,
	}
	addVoiceTrainFlags(cap, &opts, true)
	addCapabilityTrainSubcommands(cap, driver.VoiceTrainMethodCatalog(),
		func(cmd *cobra.Command, args []string) error {
			opts.Method = driver.VoiceTrainMethodRVC
			return runVoiceTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.VoiceCloning), opts, clix.JoinArgs(args))
		},
		func(info driver.TrainMethodInfo) func(*cobra.Command, []string) error {
			method, _ := driver.ParseVoiceTrainMethod(info.Method)
			return func(cmd *cobra.Command, args []string) error {
				opts.Method = method
				return runVoiceTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.VoiceCloning), opts, clix.JoinArgs(args))
			}
		},
	)
	return cap
}

func parseCapabilityType(s string) (capability.Type, error) {
	if s == "" {
		return "", fmt.Errorf("capability type is required")
	}
	cap, err := capability.Parse(s)
	if err != nil {
		return "", fmt.Errorf("unknown capability %q (valid: text, image, video, audio, mesh, voice)", s)
	}
	for _, c := range capability.All() {
		if c == capability.Training || c == capability.DatasetMgmt {
			continue
		}
		if c == cap {
			return cap, nil
		}
	}
	return "", fmt.Errorf("unknown capability %q (valid: text, image, video, audio, mesh, voice)", s)
}
