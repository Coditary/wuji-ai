package core

import (
	"context"
	"fmt"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
	"github.com/coditary/wuji/internal/driver/dummy"
)

// Config holds core initialization options.
type Config struct {
	DefaultDriverID string
}

// Core is the central orchestrator for all driver operations.
type Core struct {
	registry      *driver.Registry
	defaultDriver string
}

// New creates a Core instance with built-in drivers registered.
func New(cfg Config) (*Core, error) {
	c := &Core{
		registry:      driver.NewRegistry(),
		defaultDriver: cfg.DefaultDriverID,
	}

	if err := c.registry.Register(dummy.New()); err != nil {
		return nil, fmt.Errorf("register dummy driver: %w", err)
	}

	if c.defaultDriver == "" {
		c.defaultDriver = dummy.DriverID
	}

	if _, err := c.registry.Get(c.defaultDriver); err != nil {
		return nil, fmt.Errorf("default driver %q: %w", c.defaultDriver, err)
	}

	return c, nil
}

func (c *Core) resolve(driverID string) (driver.Driver, error) {
	if driverID == "" {
		driverID = c.defaultDriver
	}
	return c.registry.Get(driverID)
}

// Registry returns the driver registry.
func (c *Core) Registry() *driver.Registry {
	return c.registry
}

// DefaultDriverID returns the configured default driver.
func (c *Core) DefaultDriverID() string {
	return c.defaultDriver
}

// ListDrivers returns metadata for all registered drivers.
func (c *Core) ListDrivers() []driver.Info {
	return c.registry.List()
}

func (c *Core) GenerateText(ctx context.Context, driverID string, req driver.TextRequest) (*driver.TextResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.TextGenerator](d, capability.TextGeneration)
	if err != nil {
		return nil, err
	}
	return gen.GenerateText(ctx, req)
}

func (c *Core) GenerateImage(ctx context.Context, driverID string, req driver.ImageRequest) (*driver.ImageResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.ImageGenerator](d, capability.ImageGeneration)
	if err != nil {
		return nil, err
	}
	return gen.GenerateImage(ctx, req)
}

func (c *Core) GenerateVideo(ctx context.Context, driverID string, req driver.VideoRequest) (*driver.VideoResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.VideoGenerator](d, capability.VideoGeneration)
	if err != nil {
		return nil, err
	}
	return gen.GenerateVideo(ctx, req)
}

func (c *Core) GenerateAudio(ctx context.Context, driverID string, req driver.AudioRequest) (*driver.AudioResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.AudioGenerator](d, capability.AudioGeneration)
	if err != nil {
		return nil, err
	}
	return gen.GenerateAudio(ctx, req)
}

func (c *Core) Generate3D(ctx context.Context, driverID string, req driver.Asset3DRequest) (*driver.Asset3DResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.Asset3DGenerator](d, capability.Asset3D)
	if err != nil {
		return nil, err
	}
	return gen.Generate3D(ctx, req)
}

func (c *Core) Synthesize(ctx context.Context, driverID string, req driver.TTSRequest) (*driver.TTSResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.SpeechSynthesizer](d, capability.TTS)
	if err != nil {
		return nil, err
	}
	return gen.Synthesize(ctx, req)
}

func (c *Core) Transcribe(ctx context.Context, driverID string, req driver.STTRequest) (*driver.STTResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.SpeechTranscriber](d, capability.STT)
	if err != nil {
		return nil, err
	}
	return gen.Transcribe(ctx, req)
}

func (c *Core) CloneVoice(ctx context.Context, driverID string, req driver.VoiceRequest) (*driver.VoiceResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.VoiceCloner](d, capability.VoiceCloning)
	if err != nil {
		return nil, err
	}
	return gen.CloneVoice(ctx, req)
}

func (c *Core) Train(ctx context.Context, driverID string, req driver.TrainRequest) (*driver.TrainResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.Trainer](d, capability.Training)
	if err != nil {
		return nil, err
	}
	return gen.Train(ctx, req)
}

func (c *Core) ManageDataset(ctx context.Context, driverID string, req driver.DatasetRequest) (*driver.DatasetResponse, error) {
	d, err := c.resolve(driverID)
	if err != nil {
		return nil, err
	}
	gen, err := driver.As[driver.DatasetManager](d, capability.DatasetMgmt)
	if err != nil {
		return nil, err
	}
	return gen.ManageDataset(ctx, req)
}

// Close shuts down the core and all registered drivers.
func (c *Core) Close() error {
	return c.registry.Close()
}
