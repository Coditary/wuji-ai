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
		"[dummy:text] %q (max_tokens=%d, temperature=%.1f)",
		req.Prompt, req.MaxTokens, req.Temperature,
	)
	return &driver.TextResponse{Text: text, TokensUsed: len(text) / 4, FinishReason: "stop"}, nil
}

func (d *Driver) GenerateImage(_ context.Context, req driver.ImageRequest) (*driver.ImageResponse, error) {
	return &driver.ImageResponse{
		Path:   fmt.Sprintf("/tmp/wuji/dummy-image-%dx%d.png", req.Width, req.Height),
		Format: "png",
	}, nil
}

func (d *Driver) GenerateVideo(_ context.Context, req driver.VideoRequest) (*driver.VideoResponse, error) {
	return &driver.VideoResponse{
		Path:     fmt.Sprintf("/tmp/wuji/dummy-video-%.0fs.mp4", req.Duration),
		Duration: req.Duration,
	}, nil
}

func (d *Driver) GenerateAudio(_ context.Context, req driver.AudioRequest) (*driver.AudioResponse, error) {
	return &driver.AudioResponse{
		Path:     fmt.Sprintf("/tmp/wuji/dummy-audio-%.0fs.wav", req.Duration),
		Duration: req.Duration,
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
	return &driver.TTSResponse{
		Path:     fmt.Sprintf("/tmp/wuji/dummy-tts-%s.wav", req.Voice),
		Duration: float32(len(req.Text)) * 0.05,
	}, nil
}

func (d *Driver) Transcribe(_ context.Context, req driver.STTRequest) (*driver.STTResponse, error) {
	return &driver.STTResponse{
		Text:       fmt.Sprintf("[dummy:stt] transcribed from %q", req.AudioPath),
		Confidence: 0.99,
	}, nil
}

func (d *Driver) CloneVoice(_ context.Context, req driver.VoiceRequest) (*driver.VoiceResponse, error) {
	return &driver.VoiceResponse{
		VoiceID: fmt.Sprintf("voice-%s", req.Name),
		Name:    req.Name,
	}, nil
}

func (d *Driver) Train(_ context.Context, req driver.TrainRequest) (*driver.TrainResponse, error) {
	return &driver.TrainResponse{
		JobID:  fmt.Sprintf("job-%s-%s", req.DatasetID, req.ModelType),
		Status: fmt.Sprintf("queued (%d epochs)", req.Epochs),
	}, nil
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
