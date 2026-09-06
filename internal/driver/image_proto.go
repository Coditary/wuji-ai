package driver

import wujiv1 "github.com/coditary/wuji/api/proto/v1"

// ImageRequestFromProto converts a protobuf generate-image request.
func ImageRequestFromProto(req *wujiv1.GenerateImageRequest) ImageRequest {
	if req == nil {
		return ImageRequest{}
	}

	out := ImageRequest{
		Prompt:            req.GetPrompt(),
		NegativePrompt:    req.GetNegativePrompt(),
		Model:             req.GetModel(),
		Width:             int(req.GetWidth()),
		Height:            int(req.GetHeight()),
		Steps:             int(req.GetSteps()),
		Sampler:           req.GetSampler(),
		CFGScale:          req.GetCfgScale(),
		BatchSize:         int(req.GetBatchSize()),
		BatchCount:        int(req.GetBatchCount()),
		DenoisingStrength: req.GetDenoisingStrength(),
		InitImagePath:     req.GetInitImagePath(),
		LoRAs:             append([]string(nil), req.GetLoras()...),
	}
	if req.Seed != nil {
		seed := int(req.GetSeed())
		out.Seed = &seed
	}
	return out
}

// ImageRequestToProto converts an ImageRequest to protobuf.
func ImageRequestToProto(req ImageRequest) *wujiv1.GenerateImageRequest {
	out := &wujiv1.GenerateImageRequest{
		Prompt:            req.Prompt,
		NegativePrompt:    req.NegativePrompt,
		Model:             req.Model,
		Width:             int32(req.Width),
		Height:            int32(req.Height),
		Steps:             int32(req.Steps),
		Sampler:           req.Sampler,
		CfgScale:          req.CFGScale,
		BatchSize:         int32(req.BatchSize),
		BatchCount:        int32(req.BatchCount),
		DenoisingStrength: req.DenoisingStrength,
		InitImagePath:     req.InitImagePath,
		Loras:             append([]string(nil), req.LoRAs...),
	}
	if req.Seed != nil {
		seed := int32(*req.Seed)
		out.Seed = &seed
	}
	return out
}

// ImageResponseFromProto converts a protobuf generate-image response.
func ImageResponseFromProto(resp *wujiv1.GenerateImageResponse) *ImageResponse {
	if resp == nil {
		return &ImageResponse{}
	}
	return &ImageResponse{
		Path:   resp.GetPath(),
		Paths:  append([]string(nil), resp.GetPaths()...),
		Format: resp.GetFormat(),
	}
}

// ImageResponseToProto converts an ImageResponse to protobuf.
func ImageResponseToProto(resp *ImageResponse) *wujiv1.GenerateImageResponse {
	if resp == nil {
		return &wujiv1.GenerateImageResponse{}
	}
	return &wujiv1.GenerateImageResponse{
		Path:   resp.Path,
		Paths:  append([]string(nil), resp.Paths...),
		Format: resp.Format,
	}
}
