package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	wujiv1 "github.com/coditary/wuji/api/proto/v1"
	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

// RemoteDriver wraps a driver exposed over gRPC.
type RemoteDriver struct {
	info   driver.Info
	client wujiv1.DriverServiceClient
	conn   *grpc.ClientConn
}

// Connect creates a remote driver client for the given endpoint.
func Connect(ctx context.Context, endpoint string) (*RemoteDriver, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to driver at %s: %w", endpoint, err)
	}

	client := wujiv1.NewDriverServiceClient(conn)

	resp, err := client.GetInfo(ctx, &wujiv1.GetInfoRequest{})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("get driver info: %w", err)
	}

	meta := resp.GetMetadata()
	caps := make([]capability.Type, 0, len(meta.GetCapabilities()))
	for _, c := range meta.GetCapabilities() {
		caps = append(caps, capability.Type(c))
	}

	return &RemoteDriver{
		info: driver.Info{
			ID:           meta.GetId(),
			Name:         meta.GetName(),
			Version:      meta.GetVersion(),
			Description:  meta.GetDescription(),
			Capabilities: caps,
			Remote:       true,
			Endpoint:     endpoint,
		},
		client: client,
		conn:   conn,
	}, nil
}

func (d *RemoteDriver) Info() driver.Info  { return d.info }
func (d *RemoteDriver) Capabilities() []capability.Type { return d.info.Capabilities }

func (d *RemoteDriver) GenerateText(ctx context.Context, req driver.TextRequest) (*driver.TextResponse, error) {
	resp, err := d.client.GenerateText(ctx, &wujiv1.GenerateTextRequest{
		Prompt: req.Prompt, MaxTokens: int32(req.MaxTokens), Temperature: req.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("remote generate text: %w", err)
	}
	return &driver.TextResponse{
		Text: resp.GetText(), TokensUsed: int(resp.GetTokensUsed()), FinishReason: resp.GetFinishReason(),
	}, nil
}

func (d *RemoteDriver) GenerateImage(ctx context.Context, req driver.ImageRequest) (*driver.ImageResponse, error) {
	resp, err := d.client.GenerateImage(ctx, &wujiv1.GenerateImageRequest{
		Prompt: req.Prompt, Width: int32(req.Width), Height: int32(req.Height), Steps: int32(req.Steps),
	})
	if err != nil {
		return nil, fmt.Errorf("remote generate image: %w", err)
	}
	return &driver.ImageResponse{Path: resp.GetPath(), Format: resp.GetFormat()}, nil
}

func (d *RemoteDriver) GenerateVideo(ctx context.Context, req driver.VideoRequest) (*driver.VideoResponse, error) {
	resp, err := d.client.GenerateVideo(ctx, &wujiv1.GenerateVideoRequest{
		Prompt: req.Prompt, Duration: req.Duration, Fps: int32(req.FPS),
	})
	if err != nil {
		return nil, fmt.Errorf("remote generate video: %w", err)
	}
	return &driver.VideoResponse{Path: resp.GetPath(), Duration: resp.GetDuration()}, nil
}

func (d *RemoteDriver) GenerateAudio(ctx context.Context, req driver.AudioRequest) (*driver.AudioResponse, error) {
	resp, err := d.client.GenerateAudio(ctx, &wujiv1.GenerateAudioRequest{
		Prompt: req.Prompt, Duration: req.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("remote generate audio: %w", err)
	}
	return &driver.AudioResponse{Path: resp.GetPath(), Duration: resp.GetDuration()}, nil
}

func (d *RemoteDriver) Generate3D(ctx context.Context, req driver.Asset3DRequest) (*driver.Asset3DResponse, error) {
	resp, err := d.client.Generate3D(ctx, &wujiv1.Generate3DRequest{
		Prompt: req.Prompt, Format: req.Format,
	})
	if err != nil {
		return nil, fmt.Errorf("remote generate 3d: %w", err)
	}
	return &driver.Asset3DResponse{Path: resp.GetPath(), Format: resp.GetFormat()}, nil
}

func (d *RemoteDriver) Synthesize(ctx context.Context, req driver.TTSRequest) (*driver.TTSResponse, error) {
	resp, err := d.client.Synthesize(ctx, &wujiv1.SynthesizeRequest{
		Text: req.Text, Voice: req.Voice,
	})
	if err != nil {
		return nil, fmt.Errorf("remote synthesize: %w", err)
	}
	return &driver.TTSResponse{Path: resp.GetPath(), Duration: resp.GetDuration()}, nil
}

func (d *RemoteDriver) Transcribe(ctx context.Context, req driver.STTRequest) (*driver.STTResponse, error) {
	resp, err := d.client.Transcribe(ctx, &wujiv1.TranscribeRequest{
		AudioPath: req.AudioPath, Language: req.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("remote transcribe: %w", err)
	}
	return &driver.STTResponse{Text: resp.GetText(), Confidence: resp.GetConfidence()}, nil
}

func (d *RemoteDriver) CloneVoice(ctx context.Context, req driver.VoiceRequest) (*driver.VoiceResponse, error) {
	resp, err := d.client.CloneVoice(ctx, &wujiv1.CloneVoiceRequest{
		SamplePath: req.SamplePath, Name: req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("remote clone voice: %w", err)
	}
	return &driver.VoiceResponse{VoiceID: resp.GetVoiceId(), Name: resp.GetName()}, nil
}

func (d *RemoteDriver) Train(ctx context.Context, req driver.TrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.Train(ctx, &wujiv1.TrainRequest{
		DatasetId: req.DatasetID, ModelType: req.ModelType, Epochs: int32(req.Epochs),
	})
	if err != nil {
		return nil, fmt.Errorf("remote train: %w", err)
	}
	return &driver.TrainResponse{JobID: resp.GetJobId(), Status: resp.GetStatus()}, nil
}

func (d *RemoteDriver) ManageDataset(ctx context.Context, req driver.DatasetRequest) (*driver.DatasetResponse, error) {
	resp, err := d.client.ManageDataset(ctx, &wujiv1.ManageDatasetRequest{
		Action: string(req.Action), Name: req.Name, Path: req.Path,
	})
	if err != nil {
		return nil, fmt.Errorf("remote manage dataset: %w", err)
	}

	entries := make([]driver.DatasetEntry, 0, len(resp.GetDatasets()))
	for _, ds := range resp.GetDatasets() {
		entries = append(entries, driver.DatasetEntry{
			ID: ds.GetId(), Name: ds.GetName(), Path: ds.GetPath(), Size: ds.GetSize(),
		})
	}
	return &driver.DatasetResponse{Datasets: entries, Message: resp.GetMessage()}, nil
}

func (d *RemoteDriver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}
