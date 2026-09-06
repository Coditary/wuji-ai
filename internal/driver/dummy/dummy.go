package dummy

import (
	"context"
	"fmt"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

const (
	DriverID   = "dummy"
	DriverName = "Dummy Driver"
)

// Driver is a placeholder backend for development and testing.
type Driver struct{}

func New() *Driver {
	return &Driver{}
}

func (d *Driver) Info() driver.Info {
	return driver.Info{
		ID:          DriverID,
		Name:        DriverName,
		Version:     "0.1.0",
		Description: "A dummy driver that returns placeholder responses for all capabilities.",
		Capabilities: []capability.Type{
			capability.TextGeneration,
			capability.ImageGeneration,
			capability.VideoGeneration,
			capability.AudioGeneration,
			capability.Asset3D,
			capability.TTS,
			capability.STT,
			capability.VoiceCloning,
			capability.Training,
			capability.DatasetMgmt,
		},
		Remote: false,
	}
}

func (d *Driver) Capabilities() []capability.Type {
	return d.Info().Capabilities
}

func (d *Driver) GenerateText(_ context.Context, req driver.TextRequest) (*driver.TextResponse, error) {
	text := fmt.Sprintf(
		"[dummy:text] %q (max_tokens=%d, temperature=%.1f, top_p=%.2f, top_k=%d, min_p=%.2f, freq_pen=%.2f, pres_pen=%.2f, rep_pen=%.2f, stops=%d, seed=%s, ctx=%d)",
		req.Prompt, req.MaxTokens, req.Temperature,
		req.TopP, req.TopK, req.MinP,
		req.FrequencyPenalty, req.PresencePenalty, req.RepetitionPenalty,
		len(req.StopSequences), formatSeed(req.Seed), req.ContextWindow,
	)
	if req.SystemPrompt != "" {
		text = fmt.Sprintf("[dummy:system] %q\n%s", req.SystemPrompt, text)
	}
	return &driver.TextResponse{Text: text, TokensUsed: len(text) / 4, FinishReason: "stop"}, nil
}

func formatSeed(seed *int) string {
	if seed == nil {
		return "random"
	}
	return fmt.Sprintf("%d", *seed)
}

func (d *Driver) GenerateImage(_ context.Context, req driver.ImageRequest) (*driver.ImageResponse, error) {
	total := req.BatchCount
	if total <= 0 {
		total = 1
	}
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	paths := make([]string, 0, total*batchSize)
	for b := 0; b < total; b++ {
		for i := 0; i < batchSize; i++ {
			idx := b*batchSize + i
			paths = append(paths, fmt.Sprintf(
				"/tmp/wuji/dummy-image-%dx%d-%d.png",
				req.Width, req.Height, idx,
			))
		}
	}

	return &driver.ImageResponse{
		Path:   paths[0],
		Paths:  paths,
		Format: "png",
	}, nil
}

func (d *Driver) GenerateVideo(_ context.Context, req driver.VideoRequest) (*driver.VideoResponse, error) {
	frames := req.Frames
	fps := req.FPS
	if fps <= 0 {
		fps = 24
	}
	if frames <= 0 && req.Duration > 0 {
		frames = int(req.Duration * float32(fps))
	}
	if frames <= 0 {
		frames = fps * 5
	}

	duration := req.EffectiveDuration()
	if duration <= 0 {
		duration = float32(frames) / float32(fps)
	}

	return &driver.VideoResponse{
		Path:     fmt.Sprintf("/tmp/wuji/dummy-video-%dfr-%dfps.mp4", frames, fps),
		Duration: duration,
		Frames:   frames,
		FPS:      fps,
	}, nil
}

func (d *Driver) GenerateAudio(_ context.Context, req driver.AudioRequest) (*driver.AudioResponse, error) {
	duration := req.Duration
	if duration <= 0 {
		duration = 10
	}
	sampleRate := req.SampleRate
	if sampleRate <= 0 {
		sampleRate = 44100
	}

	return &driver.AudioResponse{
		Path:       fmt.Sprintf("/tmp/wuji/dummy-audio-%.0fs-%dhz.wav", duration, sampleRate),
		Duration:   duration,
		SampleRate: sampleRate,
		Format:     "wav",
	}, nil
}

func (d *Driver) Generate3D(_ context.Context, req driver.Asset3DRequest) (*driver.Asset3DResponse, error) {
	format := req.Format
	if format == "" {
		format = "glb"
	}
	return &driver.Asset3DResponse{
		Path:   fmt.Sprintf("/tmp/wuji/dummy-asset.%s", format),
		Format: format,
	}, nil
}

func (d *Driver) Synthesize(_ context.Context, req driver.TTSRequest) (*driver.TTSResponse, error) {
	voice := req.Voice
	if voice == "" {
		voice = "default"
	}
	speed := req.Speed
	if speed <= 0 {
		speed = 1.0
	}
	duration := float32(len(req.Text)) * 0.05 / speed
	return &driver.TTSResponse{
		Path:       fmt.Sprintf("/tmp/wuji/dummy-tts-%s.wav", voice),
		Duration:   duration,
		SampleRate: 22050,
		Format:     "wav",
	}, nil
}

func (d *Driver) Transcribe(_ context.Context, req driver.STTRequest) (*driver.STTResponse, error) {
	text := fmt.Sprintf("[dummy:stt] transcribed from %q", req.AudioPath)
	resp := &driver.STTResponse{
		Text:       text,
		Confidence: 0.99,
	}
	if req.WordTimestamps {
		resp.Words = []driver.WordTimestamp{
			{Word: "[dummy:stt]", Start: 0, End: 0.4},
			{Word: "transcribed", Start: 0.4, End: 1.0},
			{Word: "from", Start: 1.0, End: 1.2},
		}
	}
	return resp, nil
}

func (d *Driver) CloneVoice(_ context.Context, req driver.VoiceRequest) (*driver.VoiceResponse, error) {
	resp := &driver.VoiceResponse{
		VoiceID: fmt.Sprintf("voice-%s", req.Name),
		Name:    req.Name,
	}
	if req.Mode == driver.VoiceCloneModeConversion && req.SourcePath != "" {
		resp.OutputPath = fmt.Sprintf("/tmp/wuji/dummy-vc-%s.wav", req.Name)
	}
	return resp, nil
}

func (d *Driver) makeTrainResp(cap capability.Type, name, dataset, suffix string, epochs int, output string) *driver.TrainResponse {
	if output == "" && name != "" {
		output = fmt.Sprintf("/tmp/wuji/trained-%s-%s", cap, name)
	}
	return &driver.TrainResponse{
		JobID:      fmt.Sprintf("job-%s-%s-%s", cap, dataset, suffix),
		Status:     fmt.Sprintf("queued %s training (%d epochs)", cap, epochs),
		Capability: cap,
		OutputPath: output,
	}
}

func (d *Driver) TrainText(_ context.Context, req driver.TextTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.TextGeneration, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) TrainImage(_ context.Context, req driver.ImageTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.ImageGeneration, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) TrainVideo(_ context.Context, req driver.VideoTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.VideoGeneration, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) TrainAudio(_ context.Context, req driver.AudioTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.AudioGeneration, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) Train3D(_ context.Context, req driver.Asset3DTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.Asset3D, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) TrainTTS(_ context.Context, req driver.TTSTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.TTS, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) TrainSTT(_ context.Context, req driver.STTTrainRequest) (*driver.TrainResponse, error) {
	return d.makeTrainResp(capability.STT, req.Name, req.DatasetID, req.BaseModel, req.Epochs, req.OutputPath), nil
}

func (d *Driver) TrainVoice(_ context.Context, req driver.VoiceTrainRequest) (*driver.TrainResponse, error) {
	suffix := req.PretrainedModel
	if suffix == "" {
		suffix = "rvc"
	}
	return d.makeTrainResp(capability.VoiceCloning, req.Name, req.DatasetID, suffix, req.Epochs, req.OutputPath), nil
}

func (d *Driver) ManageDataset(_ context.Context, req driver.DatasetRequest) (*driver.DatasetResponse, error) {
	switch req.Action {
	case driver.DatasetList:
		return &driver.DatasetResponse{
			Datasets: []driver.DatasetEntry{
				{ID: "ds-1", Name: "example", Path: "/data/example", Size: 1024},
			},
		}, nil
	case driver.DatasetCreate:
		return &driver.DatasetResponse{
			Message: fmt.Sprintf("created dataset %q at %s", req.Name, req.Path),
		}, nil
	case driver.DatasetDelete:
		return &driver.DatasetResponse{
			Message: fmt.Sprintf("deleted dataset %q", req.Name),
		}, nil
	default:
		return nil, fmt.Errorf("unknown dataset action: %s", req.Action)
	}
}

func (d *Driver) Close() error {
	return nil
}
