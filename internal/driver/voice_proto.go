package driver

import wujiv1 "github.com/coditary/wuji/api/proto/v1"

// VoiceRequestFromProto converts a protobuf clone-voice request.
func VoiceRequestFromProto(req *wujiv1.CloneVoiceRequest) VoiceRequest {
	if req == nil {
		return VoiceRequest{}
	}
	return VoiceRequest{
		Name:            req.GetName(),
		SamplePath:      req.GetSamplePath(),
		TargetModel:     req.GetTargetModel(),
		SourcePath:      req.GetSourcePath(),
		Mode:            VoiceCloneMode(req.GetMode()),
		PitchShift:      int(req.GetPitchShift()),
		IndexRate:       req.GetIndexRate(),
		Protect:         req.GetProtect(),
		Vocoder:         req.GetVocoder(),
		Denoise:         req.GetDenoise(),
		DenoiseStrength: req.GetDenoiseStrength(),
		ChunkSize:       int(req.GetChunkSize()),
		Crossfade:       req.GetCrossfade(),
	}
}

// VoiceRequestToProto converts a VoiceRequest to protobuf.
func VoiceRequestToProto(req VoiceRequest) *wujiv1.CloneVoiceRequest {
	return &wujiv1.CloneVoiceRequest{
		Name:            req.Name,
		SamplePath:      req.SamplePath,
		TargetModel:     req.TargetModel,
		SourcePath:      req.SourcePath,
		Mode:            string(req.Mode),
		PitchShift:      int32(req.PitchShift),
		IndexRate:       req.IndexRate,
		Protect:         req.Protect,
		Vocoder:         req.Vocoder,
		Denoise:         req.Denoise,
		DenoiseStrength: req.DenoiseStrength,
		ChunkSize:       int32(req.ChunkSize),
		Crossfade:       req.Crossfade,
	}
}

// VoiceResponseFromProto converts a protobuf clone-voice response.
func VoiceResponseFromProto(resp *wujiv1.CloneVoiceResponse) *VoiceResponse {
	if resp == nil {
		return &VoiceResponse{}
	}
	return &VoiceResponse{
		VoiceID:    resp.GetVoiceId(),
		Name:       resp.GetName(),
		OutputPath: resp.GetOutputPath(),
	}
}

// VoiceResponseToProto converts a VoiceResponse to protobuf.
func VoiceResponseToProto(resp *VoiceResponse) *wujiv1.CloneVoiceResponse {
	if resp == nil {
		return &wujiv1.CloneVoiceResponse{}
	}
	return &wujiv1.CloneVoiceResponse{
		VoiceId:    resp.VoiceID,
		Name:       resp.Name,
		OutputPath: resp.OutputPath,
	}
}
