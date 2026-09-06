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
	resp, err := d.client.GenerateText(ctx, driver.TextRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote generate text: %w", err)
	}
	return &driver.TextResponse{
		Text: resp.GetText(), TokensUsed: int(resp.GetTokensUsed()), FinishReason: resp.GetFinishReason(),
	}, nil
}

func (d *RemoteDriver) GenerateImage(ctx context.Context, req driver.ImageRequest) (*driver.ImageResponse, error) {
	resp, err := d.client.GenerateImage(ctx, driver.ImageRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote generate image: %w", err)
	}
	return driver.ImageResponseFromProto(resp), nil
}

func (d *RemoteDriver) GenerateVideo(ctx context.Context, req driver.VideoRequest) (*driver.VideoResponse, error) {
	resp, err := d.client.GenerateVideo(ctx, driver.VideoRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote generate video: %w", err)
	}
	return driver.VideoResponseFromProto(resp), nil
}

func (d *RemoteDriver) GenerateAudio(ctx context.Context, req driver.AudioRequest) (*driver.AudioResponse, error) {
	resp, err := d.client.GenerateAudio(ctx, driver.AudioRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote generate audio: %w", err)
	}
	return driver.AudioResponseFromProto(resp), nil
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
	resp, err := d.client.Synthesize(ctx, driver.TTSRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote synthesize: %w", err)
	}
	return driver.TTSResponseFromProto(resp), nil
}

func (d *RemoteDriver) Transcribe(ctx context.Context, req driver.STTRequest) (*driver.STTResponse, error) {
	resp, err := d.client.Transcribe(ctx, driver.STTRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote transcribe: %w", err)
	}
	return driver.STTResponseFromProto(resp), nil
}

func (d *RemoteDriver) CloneVoice(ctx context.Context, req driver.VoiceRequest) (*driver.VoiceResponse, error) {
	resp, err := d.client.CloneVoice(ctx, driver.VoiceRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote clone voice: %w", err)
	}
	return driver.VoiceResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainText(ctx context.Context, req driver.TextTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainText(ctx, driver.TextTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train text: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainImage(ctx context.Context, req driver.ImageTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainImage(ctx, driver.ImageTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train image: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainVideo(ctx context.Context, req driver.VideoTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainVideo(ctx, driver.VideoTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train video: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainAudio(ctx context.Context, req driver.AudioTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainAudio(ctx, driver.AudioTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train audio: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) Train3D(ctx context.Context, req driver.Asset3DTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.Train3D(ctx, driver.Asset3DTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train 3d: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainTTS(ctx context.Context, req driver.TTSTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainTTS(ctx, driver.TTSTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train tts: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainSTT(ctx context.Context, req driver.STTTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainSTT(ctx, driver.STTTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train stt: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
}

func (d *RemoteDriver) TrainVoice(ctx context.Context, req driver.VoiceTrainRequest) (*driver.TrainResponse, error) {
	resp, err := d.client.TrainVoice(ctx, driver.VoiceTrainRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("remote train voice: %w", err)
	}
	return driver.TrainResponseFromProto(resp), nil
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
