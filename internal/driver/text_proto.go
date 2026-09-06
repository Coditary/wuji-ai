package driver

import wujiv1 "github.com/coditary/wuji/api/proto/v1"

// TextRequestFromProto converts a protobuf generate-text request.
func TextRequestFromProto(req *wujiv1.GenerateTextRequest) TextRequest {
	if req == nil {
		return TextRequest{}
	}

	out := TextRequest{
		Prompt:            req.GetPrompt(),
		Model:             req.GetModel(),
		SystemPrompt:      req.GetSystemPrompt(),
		MaxTokens:         int(req.GetMaxTokens()),
		Temperature:       req.GetTemperature(),
		TopP:              req.GetTopP(),
		TopK:              int(req.GetTopK()),
		MinP:              req.GetMinP(),
		FrequencyPenalty:  req.GetFrequencyPenalty(),
		PresencePenalty:   req.GetPresencePenalty(),
		RepetitionPenalty: req.GetRepetitionPenalty(),
		StopSequences:     append([]string(nil), req.GetStopSequences()...),
		ContextWindow:     int(req.GetContextWindow()),
	}
	if req.Seed != nil {
		seed := int(req.GetSeed())
		out.Seed = &seed
	}
	return out
}

// TextRequestToProto converts a TextRequest to protobuf.
func TextRequestToProto(req TextRequest) *wujiv1.GenerateTextRequest {
	out := &wujiv1.GenerateTextRequest{
		Prompt:            req.Prompt,
		Model:             req.Model,
		SystemPrompt:      req.SystemPrompt,
		MaxTokens:         int32(req.MaxTokens),
		Temperature:       req.Temperature,
		TopP:              req.TopP,
		TopK:              int32(req.TopK),
		MinP:              req.MinP,
		FrequencyPenalty:  req.FrequencyPenalty,
		PresencePenalty:   req.PresencePenalty,
		RepetitionPenalty: req.RepetitionPenalty,
		StopSequences:     append([]string(nil), req.StopSequences...),
		ContextWindow:     int32(req.ContextWindow),
	}
	if req.Seed != nil {
		seed := int32(*req.Seed)
		out.Seed = &seed
	}
	return out
}
