package driver

import wujiv1 "github.com/coditary/wuji/api/proto/v1"

// TTSRequestFromProto converts a protobuf synthesize request.
func TTSRequestFromProto(req *wujiv1.SynthesizeRequest) TTSRequest {
	if req == nil {
		return TTSRequest{}
	}

	out := TTSRequest{
		Text:              req.GetText(),
		Voice:             req.GetVoice(),
		VoiceSamplePath:   req.GetVoiceSamplePath(),
		Language:          req.GetLanguage(),
		Speed:             req.GetSpeed(),
		Emotion:           req.GetEmotion(),
		Temperature:       req.GetTemperature(),
		RepetitionPenalty: req.GetRepetitionPenalty(),
		Model:             req.GetModel(),
	}
	if req.Seed != nil {
		seed := int(req.GetSeed())
		out.Seed = &seed
	}
	return out
}

// TTSRequestToProto converts a TTSRequest to protobuf.
func TTSRequestToProto(req TTSRequest) *wujiv1.SynthesizeRequest {
	out := &wujiv1.SynthesizeRequest{
		Text:              req.Text,
		Voice:             req.Voice,
		VoiceSamplePath:   req.VoiceSamplePath,
		Language:          req.Language,
		Speed:             req.Speed,
		Emotion:           req.Emotion,
		Temperature:       req.Temperature,
		RepetitionPenalty: req.RepetitionPenalty,
		Model:             req.Model,
	}
	if req.Seed != nil {
		seed := int32(*req.Seed)
		out.Seed = &seed
	}
	return out
}

// TTSResponseFromProto converts a protobuf synthesize response.
func TTSResponseFromProto(resp *wujiv1.SynthesizeResponse) *TTSResponse {
	if resp == nil {
		return &TTSResponse{}
	}
	return &TTSResponse{
		Path:       resp.GetPath(),
		Duration:   resp.GetDuration(),
		SampleRate: int(resp.GetSampleRate()),
		Format:     resp.GetFormat(),
	}
}

// TTSResponseToProto converts a TTSResponse to protobuf.
func TTSResponseToProto(resp *TTSResponse) *wujiv1.SynthesizeResponse {
	if resp == nil {
		return &wujiv1.SynthesizeResponse{}
	}
	return &wujiv1.SynthesizeResponse{
		Path:       resp.Path,
		Duration:   resp.Duration,
		SampleRate: int32(resp.SampleRate),
		Format:     resp.Format,
	}
}

// STTRequestFromProto converts a protobuf transcribe request.
func STTRequestFromProto(req *wujiv1.TranscribeRequest) STTRequest {
	if req == nil {
		return STTRequest{}
	}
	return STTRequest{
		AudioPath:      req.GetAudioPath(),
		Language:       req.GetLanguage(),
		Model:          req.GetModel(),
		Task:           STTTask(req.GetTask()),
		BeamSize:       int(req.GetBeamSize()),
		WordTimestamps: req.GetWordTimestamps(),
		VADEnabled:     req.GetVadEnabled(),
		VADThreshold:   req.GetVadThreshold(),
	}
}

// STTRequestToProto converts an STTRequest to protobuf.
func STTRequestToProto(req STTRequest) *wujiv1.TranscribeRequest {
	return &wujiv1.TranscribeRequest{
		AudioPath:      req.AudioPath,
		Language:       req.Language,
		Model:          req.Model,
		Task:           string(req.Task),
		BeamSize:       int32(req.BeamSize),
		WordTimestamps: req.WordTimestamps,
		VadEnabled:     req.VADEnabled,
		VadThreshold:   req.VADThreshold,
	}
}

func wordTimestampsFromProto(words []*wujiv1.WordTimestamp) []WordTimestamp {
	if len(words) == 0 {
		return nil
	}
	out := make([]WordTimestamp, 0, len(words))
	for _, w := range words {
		out = append(out, WordTimestamp{
			Word:  w.GetWord(),
			Start: w.GetStart(),
			End:   w.GetEnd(),
		})
	}
	return out
}

func wordTimestampsToProto(words []WordTimestamp) []*wujiv1.WordTimestamp {
	if len(words) == 0 {
		return nil
	}
	out := make([]*wujiv1.WordTimestamp, 0, len(words))
	for _, w := range words {
		out = append(out, &wujiv1.WordTimestamp{
			Word:  w.Word,
			Start: w.Start,
			End:   w.End,
		})
	}
	return out
}

// STTResponseFromProto converts a protobuf transcribe response.
func STTResponseFromProto(resp *wujiv1.TranscribeResponse) *STTResponse {
	if resp == nil {
		return &STTResponse{}
	}
	return &STTResponse{
		Text:       resp.GetText(),
		Confidence: resp.GetConfidence(),
		Words:      wordTimestampsFromProto(resp.GetWords()),
	}
}

// STTResponseToProto converts an STTResponse to protobuf.
func STTResponseToProto(resp *STTResponse) *wujiv1.TranscribeResponse {
	if resp == nil {
		return &wujiv1.TranscribeResponse{}
	}
	return &wujiv1.TranscribeResponse{
		Text:       resp.Text,
		Confidence: resp.Confidence,
		Words:      wordTimestampsToProto(resp.Words),
	}
}
