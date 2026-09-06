package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

type trainBase struct {
	Enabled bool
	Dataset string
	Epochs  int
	Output  string
}

func (b trainBase) validate() error {
	if b.Dataset == "" {
		return fmt.Errorf("--dataset is required with --train")
	}
	return nil
}

func addTrainBaseFlags(cmd *cobra.Command, b *trainBase) {
	cmd.Flags().BoolVar(&b.Enabled, "train", false, "train a model for this capability instead of generating")
	cmd.Flags().StringVar(&b.Dataset, "dataset", "", "training dataset ID or path (required with --train)")
	cmd.Flags().IntVar(&b.Epochs, "epochs", 10, "training epochs")
	cmd.Flags().StringVar(&b.Output, "output", "", "output path for trained weights")
}

func requirePromptUnlessTrain(train bool, args []string) error {
	if train || len(args) >= 1 {
		return nil
	}
	return fmt.Errorf("prompt required unless --train is set")
}

func printTrainResp(resp *driver.TrainResponse) {
	fmt.Fprintf(os.Stdout, "training job %s [%s]: %s\n", resp.JobID, resp.Capability, resp.Status)
	if resp.OutputPath != "" {
		fmt.Fprintf(os.Stdout, "output: %s\n", resp.OutputPath)
	}
}

type textTrainOpts struct {
	trainBase
	BaseModel     string
	LearningRate  float32
	LoRARank      int
	ContextLength int
	BatchSize     int
}

func addTextTrainFlags(cmd *cobra.Command, o *textTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base checkpoint to fine-tune")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate (0 = backend default)")
	cmd.Flags().IntVar(&o.LoRARank, "lora-rank", 0, "LoRA rank (0 = backend default)")
	cmd.Flags().IntVar(&o.ContextLength, "context-length", 0, "training context length")
	cmd.Flags().IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
}

func runTextTrain(ctx context.Context, app *App, driverID string, o textTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	resp, err := app.Core.TrainText(ctx, driverID, driver.TextTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, LoRARank: o.LoRARank,
		ContextLength: o.ContextLength, BatchSize: o.BatchSize,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type imageTrainOpts struct {
	trainBase
	BaseModel    string
	LearningRate float32
	LoRARank     int
	Width        int
	Height       int
	ClassToken   string
}

func addImageTrainFlags(cmd *cobra.Command, o *imageTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base diffusion checkpoint")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	cmd.Flags().IntVar(&o.LoRARank, "lora-rank", 0, "LoRA rank")
	cmd.Flags().IntVar(&o.Width, "train-width", 0, "training resolution width")
	cmd.Flags().IntVar(&o.Height, "train-height", 0, "training resolution height")
	cmd.Flags().StringVar(&o.ClassToken, "class-token", "", "DreamBooth class token")
}

func runImageTrain(ctx context.Context, app *App, driverID string, o imageTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	resp, err := app.Core.TrainImage(ctx, driverID, driver.ImageTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, LoRARank: o.LoRARank,
		Width: o.Width, Height: o.Height, ClassToken: o.ClassToken,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type videoTrainOpts struct {
	trainBase
	BaseModel     string
	LearningRate  float32
	Frames        int
	FPS           int
	ContextLength int
}

func addVideoTrainFlags(cmd *cobra.Command, o *videoTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base video model checkpoint")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	cmd.Flags().IntVar(&o.Frames, "train-frames", 0, "training clip length in frames")
	cmd.Flags().IntVar(&o.FPS, "train-fps", 0, "training frame rate")
	cmd.Flags().IntVar(&o.ContextLength, "context-length", 0, "temporal context window")
}

func runVideoTrain(ctx context.Context, app *App, driverID string, o videoTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	resp, err := app.Core.TrainVideo(ctx, driverID, driver.VideoTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate,
		Frames: o.Frames, FPS: o.FPS, ContextLength: o.ContextLength,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type audioTrainOpts struct {
	trainBase
	BaseModel    string
	LearningRate float32
	SampleRate   int
	TaskType     string
}

func addAudioTrainFlags(cmd *cobra.Command, o *audioTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base audio model checkpoint")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	cmd.Flags().IntVar(&o.SampleRate, "sample-rate", 0, "training sample rate in Hz")
	cmd.Flags().StringVar(&o.TaskType, "train-task", "", "training task (text-to-music, text-to-sfx, …)")
}

func runAudioTrain(ctx context.Context, app *App, driverID string, o audioTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	resp, err := app.Core.TrainAudio(ctx, driverID, driver.AudioTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate,
		SampleRate: o.SampleRate, TaskType: driver.AudioTaskType(o.TaskType),
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type asset3DTrainOpts struct {
	trainBase
	BaseModel    string
	LearningRate float32
	Format       string
}

func add3DTrainFlags(cmd *cobra.Command, o *asset3DTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base 3D model checkpoint")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	cmd.Flags().StringVar(&o.Format, "train-format", "", "output format for trained weights")
}

func run3DTrain(ctx context.Context, app *App, driverID string, o asset3DTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	format := o.Format
	resp, err := app.Core.Train3D(ctx, driverID, driver.Asset3DTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate, Format: format,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type ttsTrainOpts struct {
	trainBase
	BaseModel       string
	LearningRate    float32
	Language        string
	ReferenceSample string
}

func addTTSTrainFlags(cmd *cobra.Command, o *ttsTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base TTS model checkpoint")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	cmd.Flags().StringVar(&o.Language, "lang", "", "target language for training")
	cmd.Flags().StringVar(&o.ReferenceSample, "reference", "", "reference voice sample path")
}

func runTTSTrain(ctx context.Context, app *App, driverID string, o ttsTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	resp, err := app.Core.TrainTTS(ctx, driverID, driver.TTSTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate,
		Language: o.Language, ReferenceSample: o.ReferenceSample,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type sttTrainOpts struct {
	trainBase
	BaseModel     string
	LearningRate  float32
	Language      string
	ModelSize     string
	FreezeEncoder bool
}

func addSTTTrainFlags(cmd *cobra.Command, o *sttTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.BaseModel, "base-model", "", "base STT model checkpoint")
	cmd.Flags().Float32Var(&o.LearningRate, "learning-rate", 0, "learning rate")
	cmd.Flags().StringVar(&o.Language, "lang", "", "training language")
	cmd.Flags().StringVar(&o.ModelSize, "model-size", "", "Whisper model size (tiny, base, large-v3, …)")
	cmd.Flags().BoolVar(&o.FreezeEncoder, "freeze-encoder", false, "freeze encoder during fine-tuning")
}

func runSTTTrain(ctx context.Context, app *App, driverID string, o sttTrainOpts, name, defaultBase string) error {
	if err := o.validate(); err != nil {
		return err
	}
	base := o.BaseModel
	if base == "" {
		base = defaultBase
	}
	resp, err := app.Core.TrainSTT(ctx, driverID, driver.STTTrainRequest{
		Name: name, DatasetID: o.Dataset, BaseModel: base, OutputPath: o.Output,
		Epochs: o.Epochs, LearningRate: o.LearningRate,
		Language: o.Language, ModelSize: o.ModelSize, FreezeEncoder: o.FreezeEncoder,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

type voiceTrainOpts struct {
	trainBase
	PretrainedModel string
	SamplePath      string
	PitchShift      int
	IndexRate       float32
	BatchSize       int
	SaveEveryEpoch  int
}

func addVoiceTrainFlags(cmd *cobra.Command, o *voiceTrainOpts) {
	addTrainBaseFlags(cmd, &o.trainBase)
	cmd.Flags().StringVar(&o.PretrainedModel, "pretrained", "", "pretrained RVC/base model path")
	cmd.Flags().StringVar(&o.SamplePath, "sample", "", "reference voice sample for training")
	cmd.Flags().IntVar(&o.PitchShift, "pitch-shift", 0, "default pitch shift for training")
	cmd.Flags().Float32Var(&o.IndexRate, "index-rate", 0, "feature retrieval rate")
	cmd.Flags().IntVar(&o.BatchSize, "batch-size", 0, "training batch size")
	cmd.Flags().IntVar(&o.SaveEveryEpoch, "save-every", 0, "save checkpoint every N epochs")
}

func runVoiceTrain(ctx context.Context, app *App, driverID string, o voiceTrainOpts, name, defaultPretrained string) error {
	if err := o.validate(); err != nil {
		return err
	}
	pretrained := o.PretrainedModel
	if pretrained == "" {
		pretrained = defaultPretrained
	}
	resp, err := app.Core.TrainVoice(ctx, driverID, driver.VoiceTrainRequest{
		Name: name, DatasetID: o.Dataset, OutputPath: o.Output, Epochs: o.Epochs,
		SamplePath: o.SamplePath, PretrainedModel: pretrained,
		PitchShift: o.PitchShift, IndexRate: o.IndexRate,
		BatchSize: o.BatchSize, SaveEveryEpoch: o.SaveEveryEpoch,
	})
	if err != nil {
		return err
	}
	printTrainResp(resp)
	return nil
}

func parseCapabilityType(s string) (capability.Type, error) {
	if s == "" {
		return "", fmt.Errorf("capability type is required")
	}
	for _, c := range capability.All() {
		if c == capability.Training || c == capability.DatasetMgmt {
			continue
		}
		if string(c) == s {
			return c, nil
		}
	}
	return "", fmt.Errorf("unknown capability %q (valid: text, image, video, audio, 3d, tts, stt, voice)", s)
}
