package driver

import wujiv1 "github.com/coditary/wuji/api/proto/v1"

// AudioRequestFromProto converts a protobuf generate-audio request.
func AudioRequestFromProto(req *wujiv1.GenerateAudioRequest) AudioRequest {
	if req == nil {
		return AudioRequest{}
	}

	out := AudioRequest{
		Prompt:         req.GetPrompt(),
		Lyrics:         req.GetLyrics(),
		NegativePrompt: req.GetNegativePrompt(),
		Model:          req.GetModel(),
		Duration:       req.GetDuration(),
		Overlap:        req.GetOverlap(),
		Temperature:    req.GetTemperature(),
		CFGScale:       req.GetCfgScale(),
		TopP:           req.GetTopP(),
		TopK:           int(req.GetTopK()),
		SampleRate:     int(req.GetSampleRate()),
		TaskType:       AudioTaskType(req.GetTaskType()),
		ReferencePath:  req.GetReferencePath(),
	}
	if req.Seed != nil {
		seed := int(req.GetSeed())
		out.Seed = &seed
	}
	return out
}

// AudioRequestToProto converts an AudioRequest to protobuf.
func AudioRequestToProto(req AudioRequest) *wujiv1.GenerateAudioRequest {
	out := &wujiv1.GenerateAudioRequest{
		Prompt:         req.Prompt,
		Lyrics:         req.Lyrics,
		NegativePrompt: req.NegativePrompt,
		Model:          req.Model,
		Duration:       req.Duration,
		Overlap:        req.Overlap,
		Temperature:    req.Temperature,
		CfgScale:       req.CFGScale,
		TopP:           req.TopP,
		TopK:           int32(req.TopK),
		SampleRate:     int32(req.SampleRate),
		TaskType:       string(req.TaskType),
		ReferencePath:  req.ReferencePath,
	}
	if req.Seed != nil {
		seed := int32(*req.Seed)
		out.Seed = &seed
	}
	return out
}

// AudioResponseFromProto converts a protobuf generate-audio response.
func AudioResponseFromProto(resp *wujiv1.GenerateAudioResponse) *AudioResponse {
	if resp == nil {
		return &AudioResponse{}
	}
	return &AudioResponse{
		Path:       resp.GetPath(),
		Duration:   resp.GetDuration(),
		SampleRate: int(resp.GetSampleRate()),
		Format:     resp.GetFormat(),
	}
}

// AudioResponseToProto converts an AudioResponse to protobuf.
func AudioResponseToProto(resp *AudioResponse) *wujiv1.GenerateAudioResponse {
	if resp == nil {
		return &wujiv1.GenerateAudioResponse{}
	}
	return &wujiv1.GenerateAudioResponse{
		Path:       resp.Path,
		Duration:   resp.Duration,
		SampleRate: int32(resp.SampleRate),
		Format:     resp.Format,
	}
}
