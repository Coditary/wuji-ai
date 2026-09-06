package driver

import wujiv1 "github.com/coditary/wuji/api/proto/v1"

// VideoRequestFromProto converts a protobuf generate-video request.
func VideoRequestFromProto(req *wujiv1.GenerateVideoRequest) VideoRequest {
	if req == nil {
		return VideoRequest{}
	}

	out := VideoRequest{
		Prompt:                req.GetPrompt(),
		Duration:              req.GetDuration(),
		FPS:                   int(req.GetFps()),
		Frames:                int(req.GetFrames()),
		MotionStrength:        req.GetMotionStrength(),
		ContextLength:         int(req.GetContextLength()),
		Sampler:               req.GetSampler(),
		Scheduler:             req.GetScheduler(),
		Interpolate:           req.GetInterpolate(),
		InterpolationStrength: req.GetInterpolationStrength(),
		Mode:                  VideoMode(req.GetMode()),
		InitImagePath:         req.GetInitImagePath(),
		CameraControl:         CameraControl(req.GetCameraControl()),
		NegativePrompt:        req.GetNegativePrompt(),
		Model:                 req.GetModel(),
	}
	if req.Seed != nil {
		seed := int(req.GetSeed())
		out.Seed = &seed
	}
	return out
}

// VideoRequestToProto converts a VideoRequest to protobuf.
func VideoRequestToProto(req VideoRequest) *wujiv1.GenerateVideoRequest {
	out := &wujiv1.GenerateVideoRequest{
		Prompt:                req.Prompt,
		Duration:              req.Duration,
		Fps:                   int32(req.FPS),
		Frames:                int32(req.Frames),
		MotionStrength:        req.MotionStrength,
		ContextLength:         int32(req.ContextLength),
		Sampler:               req.Sampler,
		Scheduler:             req.Scheduler,
		Interpolate:           req.Interpolate,
		InterpolationStrength: req.InterpolationStrength,
		Mode:                  string(req.Mode),
		InitImagePath:         req.InitImagePath,
		CameraControl:         string(req.CameraControl),
		NegativePrompt:        req.NegativePrompt,
		Model:                 req.Model,
	}
	if req.Seed != nil {
		seed := int32(*req.Seed)
		out.Seed = &seed
	}
	return out
}

// VideoResponseFromProto converts a protobuf generate-video response.
func VideoResponseFromProto(resp *wujiv1.GenerateVideoResponse) *VideoResponse {
	if resp == nil {
		return &VideoResponse{}
	}
	return &VideoResponse{
		Path:     resp.GetPath(),
		Duration: resp.GetDuration(),
		Frames:   int(resp.GetFrames()),
		FPS:      int(resp.GetFps()),
	}
}

// VideoResponseToProto converts a VideoResponse to protobuf.
func VideoResponseToProto(resp *VideoResponse) *wujiv1.GenerateVideoResponse {
	if resp == nil {
		return &wujiv1.GenerateVideoResponse{}
	}
	return &wujiv1.GenerateVideoResponse{
		Path:     resp.Path,
		Duration: resp.Duration,
		Frames:   int32(resp.Frames),
		Fps:      int32(resp.FPS),
	}
}
