package cli

import (
	"fmt"

	"github.com/coditary/wuji-core/pkg/driver"
)

func buildVideoRequest(prompt string, videoTrainEnabled bool, fields videoRequestFields) (driver.VideoRequest, error) {
	if videoTrainEnabled {
		return driver.VideoRequest{Prompt: prompt, Model: fields.model}, nil
	}

	task, err := driver.InferVideoTask(driver.VideoTaskInputs{
		ImagePath: fields.imagePath,
		VideoPath: fields.videoPath,
		Scale:     fields.scale,
		ScaleSet:  fields.scaleSet,
		FPSSet:    fields.fpsSet,
	})
	if err != nil {
		return driver.VideoRequest{}, err
	}

	camera, err := driver.ParseCameraControl(fields.cameraControl)
	if err != nil {
		return driver.VideoRequest{}, err
	}
	if camera != "" && task != driver.VideoTaskGenerate {
		return driver.VideoRequest{}, fmt.Errorf("--camera is only valid for text-to-video (generate), not %q", task)
	}

	req := driver.VideoRequest{
		Task:                  task,
		Prompt:                prompt,
		Duration:              fields.duration,
		FPS:                   fields.fps,
		Frames:                fields.frames,
		MotionStrength:        fields.motionStrength,
		ContextLength:         fields.contextLength,
		Sampler:               fields.sampler,
		Scheduler:     fields.scheduler,
		InitImagePath: fields.imagePath,
		InitVideoPath:         fields.videoPath,
		CameraControl:         camera,
		Scale:                 fields.scale,
		NegativePrompt:        fields.negativePrompt,
		Model:                 fields.model,
	}
	if fields.seed >= 0 {
		req.Seed = &fields.seed
	}
	if err := req.Validate(); err != nil {
		return driver.VideoRequest{}, err
	}
	return req, nil
}

type videoRequestFields struct {
	duration              float32
	fps                   int
	frames                int
	fpsSet                bool
	motionStrength        float32
	contextLength         int
	sampler               string
	scheduler string
	imagePath string
	videoPath             string
	cameraControl         string
	negativePrompt        string
	model                 string
	seed                  int
	scale                 float32
	scaleSet              bool
}
