package cli

import (
	"strings"

	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/coditary/wuji-core/pkg/scalefactor"
)

func parseImageScaleInput(raw string, upscaleFlag bool) (float32, bool, error) {
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

func buildImageRequest(prompt string, imageTrainEnabled bool, fields imageRequestFields) (driver.ImageRequest, error) {
	if imageTrainEnabled {
		return driver.ImageRequest{Prompt: prompt, Model: fields.model}, nil
	}

	task := driver.InferImageTask(driver.ImageTaskInputs{
		Prompt:               prompt,
		InitImagePath:        fields.initImage,
		MaskImagePath:        fields.maskImage,
		ControlImagePath:     fields.controlImage,
		ControlType:          fields.controlType,
		ControlUnits:         fields.controlUnits,
		StyleImagePath:       fields.styleImage,
		Scale:                fields.scale,
		UpscaleRequested:     fields.upscaleRequested,
		EditRequested:        fields.editRequested,
		SpriteRequested: fields.spriteRequested,
	})

	units, err := driver.MergeControlUnits(fields.controlUnits, fields.controlImage, fields.controlType)
	if err != nil {
		return driver.ImageRequest{}, err
	}

	batchSize, batchCount, err := resolveImageBatch(imageBatchOptions{
		total:           fields.batchTotal,
		size:            fields.batchSize,
		count:           fields.batchCount,
		sizeSet:         fields.batchSizeSet,
		countSet:        fields.batchCountSet,
		width:           fields.width,
		height:          fields.height,
		controlUnits:    len(units),
		availableVRAMMB: fields.availableVRAMMB,
	}, task)
	if err != nil {
		return driver.ImageRequest{}, err
	}

	var mode driver.AssetMode
	if fields.mode != "" {
		mode, err = driver.ParseImageMode(fields.mode)
		if err != nil {
			return driver.ImageRequest{}, err
		}
	}

	frameWidth := fields.frameWidth
	frameHeight := fields.frameHeight
	if task == driver.ImageTaskSprite {
		if frameWidth <= 0 {
			frameWidth = fields.width
		}
		if frameHeight <= 0 {
			frameHeight = fields.height
		}
	}

	var spriteAction driver.SpriteAction
	if fields.spriteAction != "" {
		spriteAction, err = driver.ParseSpriteAction(fields.spriteAction)
		if err != nil {
			return driver.ImageRequest{}, err
		}
	}
	var spriteView driver.SpriteView
	if fields.spriteView != "" {
		spriteView, err = driver.ParseSpriteView(fields.spriteView)
		if err != nil {
			return driver.ImageRequest{}, err
		}
	}
	if fields.spriteDirections != 0 {
		if _, err = driver.ParseSpriteDirections(fields.spriteDirections); err != nil {
			return driver.ImageRequest{}, err
		}
	}

	req := driver.ImageRequest{
		Task:              task,
		Prompt:            prompt,
		NegativePrompt:    fields.negativePrompt,
		Model:             fields.model,
		Width:             fields.width,
		Height:            fields.height,
		Steps:             fields.steps,
		Sampler:           fields.sampler,
		CFGScale:          fields.cfgScale,
		BatchSize:         batchSize,
		BatchCount:        batchCount,
		DenoisingStrength: fields.denoisingStrength,
		InitImagePath:     fields.initImage,
		MaskImagePath:     fields.maskImage,
		ControlImagePath:  fields.controlImage,
		ControlUnits:      units,
		ControlMode:       fields.controlMode,
		StyleImagePath:    fields.styleImage,
		StyleWeight:       fields.styleWeight,
		Scale:             fields.scale,
		LoRAs:             fields.loras,
		Mode:              mode,
		FrameWidth:         frameWidth,
		FrameHeight:        frameHeight,
		Columns:            fields.columns,
		Rows:               fields.rows,
		FrameCount:         fields.frameCount,
		ReferenceImagePath: fields.referenceImage,
		SpriteAction:       spriteAction,
		SpriteView:         spriteView,
		SpriteDirections:   fields.spriteDirections,
		SpriteLoop:         fields.spriteLoop,
		SpritePadding:      fields.spritePadding,
		SpriteTransparent:  fields.spriteTransparent,
	}
	if task == driver.ImageTaskSprite && req.Columns <= 0 && req.Rows <= 0 && req.FrameCount <= 0 && batchSize*batchCount > 0 {
		req.FrameCount = batchSize * batchCount
		req.Columns = 4
		req.Rows = (req.FrameCount + req.Columns - 1) / req.Columns
	}
	if fields.controlType != "" && req.ControlType == "" {
		ct, err := driver.ParseImageControlType(fields.controlType)
		if err != nil {
			return driver.ImageRequest{}, err
		}
		req.ControlType = ct
	}
	if fields.seed >= 0 {
		req.Seed = &fields.seed
	}
	if err := req.Validate(); err != nil {
		return driver.ImageRequest{}, err
	}
	return req, nil
}

type imageRequestFields struct {
	negativePrompt       string
	model                string
	mode                 string
	width                int
	height               int
	steps                int
	sampler              string
	cfgScale             float32
	batchTotal           int
	batchSize            int
	batchCount           int
	batchSizeSet         bool
	batchCountSet        bool
	availableVRAMMB      int
	seed                 int
	denoisingStrength    float32
	initImage            string
	maskImage            string
	controlImage         string
	controlType          string
	controlUnits         []driver.ControlNetUnit
	controlMode          driver.ControlNetMode
	styleImage           string
	styleWeight          float32
	scale                float32
	upscaleRequested     bool
	editRequested    bool
	spriteRequested  bool
	referenceImage   string
	spriteAction     string
	spriteView       string
	spriteDirections int
	spriteLoop       bool
	spritePadding    int
	spriteTransparent bool
	frameWidth           int
	frameHeight          int
	columns              int
	rows                 int
	frameCount           int
	loras                []driver.LoRARef
}
