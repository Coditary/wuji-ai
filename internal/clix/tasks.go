package clix

import (
	"fmt"
	"strings"

	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/coditary/wuji-core/pkg/scalefactor"
)

func ParseImageScaleInput(raw string, upscaleFlag bool) (float32, bool, error) {
	if strings.TrimSpace(raw) == "" {
		if upscaleFlag {
			return 4, true, nil
		}
		return 0, upscaleFlag, nil
	}
	v, err := scalefactor.Parse(raw)
	if err != nil {
		return 0, upscaleFlag, err
	}
	return v, true, nil
}

func BuildImageRequest(prompt string, imageTrainEnabled bool, fields ImageRequestFields) (driver.ImageRequest, error) {
	if imageTrainEnabled {
		return driver.ImageRequest{Prompt: prompt, Model: fields.Model}, nil
	}

	task := driver.InferImageTask(driver.ImageTaskInputs{
		Prompt:           prompt,
		InitImagePath:    fields.InitImage,
		MaskImagePath:    fields.MaskImage,
		ControlImagePath: fields.ControlImage,
		ControlType:      fields.ControlType,
		ControlUnits:     fields.ControlUnits,
		StyleImagePath:   fields.StyleImage,
		Scale:            fields.Scale,
		UpscaleRequested: fields.UpscaleRequested,
		EditRequested:    fields.EditRequested,
		SpriteRequested:  fields.SpriteRequested,
	})

	units, err := driver.MergeControlUnits(fields.ControlUnits, fields.ControlImage, fields.ControlType)
	if err != nil {
		return driver.ImageRequest{}, err
	}

	batchSize, batchCount, err := ResolveImageBatch(ImageBatchOptions{
		Total:           fields.BatchTotal,
		Size:            fields.BatchSize,
		Count:           fields.BatchCount,
		SizeSet:         fields.BatchSizeSet,
		CountSet:        fields.BatchCountSet,
		Width:           fields.Width,
		Height:          fields.Height,
		ControlUnits:    len(units),
		AvailableVRAMMB: fields.AvailableVRAMMB,
	}, task)
	if err != nil {
		return driver.ImageRequest{}, err
	}

	var mode driver.AssetMode
	if fields.Mode != "" {
		mode, err = driver.ParseImageMode(fields.Mode)
		if err != nil {
			return driver.ImageRequest{}, err
		}
	}

	frameWidth := fields.FrameWidth
	frameHeight := fields.FrameHeight
	if task == driver.ImageTaskSprite {
		if frameWidth <= 0 {
			frameWidth = fields.Width
		}
		if frameHeight <= 0 {
			frameHeight = fields.Height
		}
	}

	var spriteAction driver.SpriteAction
	if fields.SpriteAction != "" {
		spriteAction, err = driver.ParseSpriteAction(fields.SpriteAction)
		if err != nil {
			return driver.ImageRequest{}, err
		}
	}
	var spriteView driver.SpriteView
	if fields.SpriteView != "" {
		spriteView, err = driver.ParseSpriteView(fields.SpriteView)
		if err != nil {
			return driver.ImageRequest{}, err
		}
	}
	if fields.SpriteDirections != 0 {
		if _, err = driver.ParseSpriteDirections(fields.SpriteDirections); err != nil {
			return driver.ImageRequest{}, err
		}
	}

	req := driver.ImageRequest{
		Task:               task,
		Prompt:             prompt,
		NegativePrompt:     fields.NegativePrompt,
		Model:              fields.Model,
		Width:              fields.Width,
		Height:             fields.Height,
		Steps:              fields.Steps,
		Sampler:            fields.Sampler,
		CFGScale:           fields.CFGScale,
		BatchSize:          batchSize,
		BatchCount:         batchCount,
		DenoisingStrength:  fields.DenoisingStrength,
		InitImagePath:      fields.InitImage,
		MaskImagePath:      fields.MaskImage,
		ControlImagePath:   fields.ControlImage,
		ControlUnits:       units,
		ControlMode:        fields.ControlMode,
		StyleImagePath:     fields.StyleImage,
		StyleWeight:        fields.StyleWeight,
		Scale:              fields.Scale,
		LoRAs:              fields.LoRAs,
		Mode:               mode,
		FrameWidth:         frameWidth,
		FrameHeight:        frameHeight,
		Columns:            fields.Columns,
		Rows:               fields.Rows,
		FrameCount:         fields.FrameCount,
		ReferenceImagePath: fields.ReferenceImage,
		SpriteAction:       spriteAction,
		SpriteView:         spriteView,
		SpriteDirections:   fields.SpriteDirections,
		SpriteLoop:         fields.SpriteLoop,
		SpritePadding:      fields.SpritePadding,
		SpriteTransparent:  fields.SpriteTransparent,
	}
	if task == driver.ImageTaskSprite && req.Columns <= 0 && req.Rows <= 0 && req.FrameCount <= 0 && batchSize*batchCount > 0 {
		req.FrameCount = batchSize * batchCount
		req.Columns = 4
		req.Rows = (req.FrameCount + req.Columns - 1) / req.Columns
	}
	if fields.ControlType != "" && req.ControlType == "" {
		ct, err := driver.ParseImageControlType(fields.ControlType)
		if err != nil {
			return driver.ImageRequest{}, err
		}
		req.ControlType = ct
	}
	if fields.Seed >= 0 {
		req.Seed = &fields.Seed
	}
	if err := req.Validate(); err != nil {
		return driver.ImageRequest{}, err
	}
	return req, nil
}

type ImageRequestFields struct {
	NegativePrompt    string
	Model             string
	Mode              string
	Width             int
	Height            int
	Steps             int
	Sampler           string
	CFGScale          float32
	BatchTotal        int
	BatchSize         int
	BatchCount        int
	BatchSizeSet      bool
	BatchCountSet     bool
	AvailableVRAMMB   int
	Seed              int
	DenoisingStrength float32
	InitImage         string
	MaskImage         string
	ControlImage      string
	ControlType       string
	ControlUnits      []driver.ControlNetUnit
	ControlMode       driver.ControlNetMode
	StyleImage        string
	StyleWeight       float32
	Scale             float32
	UpscaleRequested  bool
	EditRequested     bool
	SpriteRequested   bool
	ReferenceImage    string
	SpriteAction      string
	SpriteView        string
	SpriteDirections  int
	SpriteLoop        bool
	SpritePadding     int
	SpriteTransparent bool
	FrameWidth        int
	FrameHeight       int
	Columns           int
	Rows              int
	FrameCount        int
	LoRAs             []driver.LoRARef
}

func BuildAudioRequest(prompt string, fields AudioRequestFields) (driver.AudioRequest, error) {
	task, err := driver.InferAudioTask(driver.AudioTaskInputs{
		ReferencePath:   fields.ReferencePath,
		SFXRequested:    fields.SFXRequested,
		SpeechRequested: fields.SpeechRequested,
		VoiceName:       fields.Voice,
	})
	if err != nil {
		return driver.AudioRequest{}, err
	}

	duration := fields.Duration
	if task == driver.AudioTaskTextToSpeech && !fields.DurationSet {
		duration = 0
	}

	req := driver.AudioRequest{
		Task: task, Prompt: prompt, Lyrics: fields.Lyrics, NegativePrompt: fields.NegativePrompt,
		Model: fields.Model, Duration: duration, Overlap: fields.Overlap, Temperature: fields.Temperature,
		CFGScale: fields.CFGScale, TopP: fields.TopP, TopK: fields.TopK, SampleRate: fields.SampleRate,
		ReferencePath: fields.ReferencePath, Voice: fields.Voice, Language: fields.Language,
		Speed: fields.Speed, Pitch: fields.Pitch, Emotion: fields.Emotion, Style: fields.Style,
		Energy: fields.Energy, Format: fields.Format,
	}
	if fields.Seed >= 0 {
		req.Seed = &fields.Seed
	}
	if err := req.Validate(); err != nil {
		return driver.AudioRequest{}, err
	}
	return req, nil
}

type AudioRequestFields struct {
	Lyrics          string
	NegativePrompt  string
	Model           string
	Duration        float32
	DurationSet     bool
	Overlap         float32
	Temperature     float32
	CFGScale        float32
	TopP            float32
	TopK            int
	SampleRate      int
	ReferencePath   string
	Voice           string
	Language        string
	Speed           float32
	Pitch           int
	Emotion         string
	Style           string
	Energy          float32
	Format          string
	SFXRequested    bool
	SpeechRequested bool
	Seed            int
}

func BuildVideoRequest(prompt string, videoTrainEnabled bool, fields VideoRequestFields) (driver.VideoRequest, error) {
	if videoTrainEnabled {
		return driver.VideoRequest{Prompt: prompt, Model: fields.Model}, nil
	}

	task, err := driver.InferVideoTask(driver.VideoTaskInputs{
		ImagePath: fields.ImagePath,
		VideoPath: fields.VideoPath,
		Scale:     fields.Scale,
		ScaleSet:  fields.ScaleSet,
		FPSSet:    fields.FPSSet,
	})
	if err != nil {
		return driver.VideoRequest{}, err
	}

	camera, err := driver.ParseCameraControl(fields.CameraControl)
	if err != nil {
		return driver.VideoRequest{}, err
	}
	if camera != "" && task != driver.VideoTaskGenerate {
		return driver.VideoRequest{}, fmt.Errorf("--camera is only valid for text-to-video (generate), not %q", task)
	}

	req := driver.VideoRequest{
		Task:           task,
		Prompt:         prompt,
		Duration:       fields.Duration,
		FPS:            fields.FPS,
		Frames:         fields.Frames,
		MotionStrength: fields.MotionStrength,
		ContextLength:  fields.ContextLength,
		Sampler:        fields.Sampler,
		Scheduler:      fields.Scheduler,
		InitImagePath:  fields.ImagePath,
		InitVideoPath:  fields.VideoPath,
		CameraControl:  camera,
		Scale:          fields.Scale,
		NegativePrompt: fields.NegativePrompt,
		Model:          fields.Model,
	}
	if fields.Seed >= 0 {
		req.Seed = &fields.Seed
	}
	if err := req.Validate(); err != nil {
		return driver.VideoRequest{}, err
	}
	return req, nil
}

type VideoRequestFields struct {
	Duration       float32
	FPS            int
	Frames         int
	FPSSet         bool
	MotionStrength float32
	ContextLength  int
	Sampler        string
	Scheduler      string
	ImagePath      string
	VideoPath      string
	CameraControl  string
	NegativePrompt string
	Model          string
	Seed           int
	Scale          float32
	ScaleSet       bool
}

func BuildMeshRequest(prompt string, meshTrainEnabled bool, fields MeshRequestFields) (driver.MeshRequest, error) {
	if meshTrainEnabled {
		return driver.MeshRequest{Prompt: prompt, Format: fields.Format}, nil
	}

	var mode driver.MeshMode
	var err error
	if fields.Mode != "" {
		mode, err = driver.ParseMeshMode(fields.Mode)
		if err != nil {
			return driver.MeshRequest{}, err
		}
	}

	var task driver.MeshTask
	if fields.TaskExplicit != "" {
		task, err = driver.ParseMeshTask(fields.TaskExplicit)
		if err != nil {
			return driver.MeshRequest{}, err
		}
	} else {
		task, err = driver.InferMeshTask(driver.MeshTaskInputs{
			Prompt:             prompt,
			MeshPath:           fields.MeshPath,
			HighMeshPath:       fields.HighMeshPath,
			ImagePath:          fields.ImagePath,
			Images:             fields.Images,
			VideoPath:          fields.VideoPath,
			DepthPath:          fields.DepthPath,
			PointCloudPath:     fields.PointCloudPath,
			SplatPath:          fields.SplatPath,
			MaskPath:           fields.MaskPath,
			StyleImagePath:     fields.StyleImagePath,
			TextureImagePath:   fields.TextureImagePath,
			AnimationPath:      fields.AnimationPath,
			TargetTris:         fields.TargetTris,
			TargetTrisSet:      fields.TargetTrisSet,
			RetopoRequested:    fields.RetopoRequested,
			RemeshRequested:    fields.RemeshRequested,
			RepairRequested:    fields.RepairRequested,
			SmoothRequested:    fields.SmoothRequested,
			RefineRequested:    fields.RefineRequested,
			SegmentRequested:   fields.SegmentRequested,
			RigRequested:       fields.RigRequested,
			AnimateRequested:   fields.AnimateRequested,
			RetargetRequested:  fields.RetargetRequested,
			UVRequested:        fields.UVRequested,
			PBRRequested:       fields.PBRRequested,
			SceneRequested:     fields.SceneRequested,
			VariationRequested: fields.VariationRequested,
			EditRequested:      fields.EditRequested,
			UpscaleRequested:   fields.UpscaleRequested,
			Scale:              fields.Scale,
			ScaleSet:           fields.ScaleSet,
		})
		if err != nil {
			return driver.MeshRequest{}, err
		}
	}

	scale := fields.Scale
	if task == driver.MeshTaskUpscale && scale <= 0 && fields.UpscaleRequested {
		scale = 2
	}

	initImage := fields.ImagePath
	initImages := append([]string(nil), fields.Images...)
	if initImage == "" && len(initImages) == 1 {
		initImage = initImages[0]
		initImages = nil
	}

	req := driver.MeshRequest{
		Task:               task,
		Prompt:             prompt,
		Format:             fields.Format,
		Model:              fields.Model,
		Mode:               mode,
		TargetTris:         fields.TargetTris,
		Scale:              scale,
		InitMeshPath:       fields.MeshPath,
		HighMeshPath:       fields.HighMeshPath,
		InitImagePath:      initImage,
		InitImagePaths:     initImages,
		InitVideoPath:      fields.VideoPath,
		InitDepthPath:      fields.DepthPath,
		InitPointCloudPath: fields.PointCloudPath,
		InitSplatPath:      fields.SplatPath,
		MaskPath:           fields.MaskPath,
		StyleImagePath:     fields.StyleImagePath,
		TextureImagePath:   fields.TextureImagePath,
		AnimationPath:      fields.AnimationPath,
	}
	if fields.Representation != "" {
		req.Representation = fields.Representation
	}
	if fields.Seed >= 0 {
		req.Seed = &fields.Seed
	}
	if err := req.Validate(); err != nil {
		return driver.MeshRequest{}, err
	}
	return req, nil
}

type MeshRequestFields struct {
	TaskExplicit       string
	Model              string
	Format             string
	Mode               string
	Seed               int
	TargetTris         int
	TargetTrisSet      bool
	Scale              float32
	ScaleSet           bool
	UpscaleRequested   bool
	MeshPath           string
	HighMeshPath       string
	ImagePath          string
	Images             []string
	VideoPath          string
	DepthPath          string
	PointCloudPath     string
	SplatPath          string
	MaskPath           string
	StyleImagePath     string
	TextureImagePath   string
	AnimationPath      string
	RetopoRequested    bool
	RemeshRequested    bool
	RepairRequested    bool
	SmoothRequested    bool
	RefineRequested    bool
	SegmentRequested   bool
	RigRequested       bool
	AnimateRequested   bool
	RetargetRequested  bool
	UVRequested        bool
	PBRRequested       bool
	SceneRequested     bool
	VariationRequested bool
	EditRequested      bool
	Representation     driver.MeshRepresentation
}
