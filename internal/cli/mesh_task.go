package cli

import (
	"github.com/coditary/wuji-core/pkg/driver"
)

type meshRequestFields struct {
	taskExplicit       string
	model              string
	format             string
	mode               string
	seed               int
	targetTris         int
	targetTrisSet      bool
	scale              float32
	scaleSet           bool
	upscaleRequested   bool
	meshPath           string
	highMeshPath       string
	imagePath          string
	images             []string
	videoPath          string
	depthPath          string
	pointCloudPath     string
	splatPath          string
	maskPath           string
	styleImagePath     string
	textureImagePath   string
	animationPath      string
	retopoRequested    bool
	remeshRequested    bool
	repairRequested    bool
	smoothRequested    bool
	refineRequested    bool
	segmentRequested   bool
	rigRequested       bool
	animateRequested   bool
	retargetRequested  bool
	uvRequested        bool
	pbrRequested       bool
	sceneRequested     bool
	variationRequested bool
	editRequested      bool
	representation     driver.MeshRepresentation
}

func buildMeshRequest(prompt string, meshTrainEnabled bool, fields meshRequestFields) (driver.MeshRequest, error) {
	if meshTrainEnabled {
		return driver.MeshRequest{Prompt: prompt, Format: fields.format}, nil
	}

	var mode driver.MeshMode
	var err error
	if fields.mode != "" {
		mode, err = driver.ParseMeshMode(fields.mode)
		if err != nil {
			return driver.MeshRequest{}, err
		}
	}

	var task driver.MeshTask
	if fields.taskExplicit != "" {
		task, err = driver.ParseMeshTask(fields.taskExplicit)
		if err != nil {
			return driver.MeshRequest{}, err
		}
	} else {
		task, err = driver.InferMeshTask(driver.MeshTaskInputs{
			Prompt:             prompt,
			MeshPath:           fields.meshPath,
			HighMeshPath:       fields.highMeshPath,
			ImagePath:          fields.imagePath,
			Images:             fields.images,
			VideoPath:          fields.videoPath,
			DepthPath:          fields.depthPath,
			PointCloudPath:     fields.pointCloudPath,
			SplatPath:          fields.splatPath,
			MaskPath:           fields.maskPath,
			StyleImagePath:     fields.styleImagePath,
			TextureImagePath:   fields.textureImagePath,
			AnimationPath:      fields.animationPath,
			TargetTris:         fields.targetTris,
			TargetTrisSet:      fields.targetTrisSet,
			RetopoRequested:    fields.retopoRequested,
			RemeshRequested:    fields.remeshRequested,
			RepairRequested:    fields.repairRequested,
			SmoothRequested:    fields.smoothRequested,
			RefineRequested:    fields.refineRequested,
			SegmentRequested:   fields.segmentRequested,
			RigRequested:       fields.rigRequested,
			AnimateRequested:   fields.animateRequested,
			RetargetRequested:  fields.retargetRequested,
			UVRequested:        fields.uvRequested,
			PBRRequested:       fields.pbrRequested,
			SceneRequested:     fields.sceneRequested,
			VariationRequested: fields.variationRequested,
			EditRequested:      fields.editRequested,
			UpscaleRequested:   fields.upscaleRequested,
			Scale:              fields.scale,
			ScaleSet:           fields.scaleSet,
		})
		if err != nil {
			return driver.MeshRequest{}, err
		}
	}

	scale := fields.scale
	if task == driver.MeshTaskUpscale && scale <= 0 && fields.upscaleRequested {
		scale = 2
	}

	initImage := fields.imagePath
	initImages := append([]string(nil), fields.images...)
	if initImage == "" && len(initImages) == 1 {
		initImage = initImages[0]
		initImages = nil
	}

	req := driver.MeshRequest{
		Task:               task,
		Prompt:             prompt,
		Format:             fields.format,
		Model:              fields.model,
		Mode:               mode,
		TargetTris:         fields.targetTris,
		Scale:              scale,
		InitMeshPath:       fields.meshPath,
		HighMeshPath:       fields.highMeshPath,
		InitImagePath:      initImage,
		InitImagePaths:     initImages,
		InitVideoPath:      fields.videoPath,
		InitDepthPath:      fields.depthPath,
		InitPointCloudPath: fields.pointCloudPath,
		InitSplatPath:      fields.splatPath,
		MaskPath:           fields.maskPath,
		StyleImagePath:     fields.styleImagePath,
		TextureImagePath:   fields.textureImagePath,
		AnimationPath:      fields.animationPath,
	}
	if fields.representation != "" {
		req.Representation = fields.representation
	}
	if fields.seed >= 0 {
		req.Seed = &fields.seed
	}
	if err := req.Validate(); err != nil {
		return driver.MeshRequest{}, err
	}
	return req, nil
}
