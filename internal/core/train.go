package core

import (
	"context"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

func (c *Core) TrainText(ctx context.Context, driverID string, req driver.TextTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.TextTrainer](d, capability.TextGeneration)
	if err != nil {
		return nil, err
	}
	return tr.TrainText(ctx, req)
}

func (c *Core) TrainImage(ctx context.Context, driverID string, req driver.ImageTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.ImageTrainer](d, capability.ImageGeneration)
	if err != nil {
		return nil, err
	}
	return tr.TrainImage(ctx, req)
}

func (c *Core) TrainVideo(ctx context.Context, driverID string, req driver.VideoTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.VideoTrainer](d, capability.VideoGeneration)
	if err != nil {
		return nil, err
	}
	return tr.TrainVideo(ctx, req)
}

func (c *Core) TrainAudio(ctx context.Context, driverID string, req driver.AudioTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.AudioTrainer](d, capability.AudioGeneration)
	if err != nil {
		return nil, err
	}
	return tr.TrainAudio(ctx, req)
}

func (c *Core) Train3D(ctx context.Context, driverID string, req driver.Asset3DTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.Asset3DTrainer](d, capability.Asset3D)
	if err != nil {
		return nil, err
	}
	return tr.Train3D(ctx, req)
}

func (c *Core) TrainTTS(ctx context.Context, driverID string, req driver.TTSTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.TTSTrainer](d, capability.TTS)
	if err != nil {
		return nil, err
	}
	return tr.TrainTTS(ctx, req)
}

func (c *Core) TrainSTT(ctx context.Context, driverID string, req driver.STTTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.STTTrainer](d, capability.STT)
	if err != nil {
		return nil, err
	}
	return tr.TrainSTT(ctx, req)
}

func (c *Core) TrainVoice(ctx context.Context, driverID string, req driver.VoiceTrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	tr, err := driver.As[driver.VoiceTrainer](d, capability.VoiceCloning)
	if err != nil {
		return nil, err
	}
	return tr.TrainVoice(ctx, req)
}
