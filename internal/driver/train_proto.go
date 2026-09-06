package driver

import (
	wujiv1 "github.com/coditary/wuji/api/proto/v1"
	"github.com/coditary/wuji/internal/capability"
)

func TextTrainRequestFromProto(req *wujiv1.TrainTextRequest) TextTrainRequest {
	if req == nil {
		return TextTrainRequest{}
	}
	return TextTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), LoRARank: int(req.GetLoraRank()),
		ContextLength: int(req.GetContextLength()), BatchSize: int(req.GetBatchSize()),
	}
}

func TextTrainRequestToProto(req TextTrainRequest) *wujiv1.TrainTextRequest {
	return &wujiv1.TrainTextRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, LoraRank: int32(req.LoRARank),
		ContextLength: int32(req.ContextLength), BatchSize: int32(req.BatchSize),
	}
}

func ImageTrainRequestFromProto(req *wujiv1.TrainImageRequest) ImageTrainRequest {
	if req == nil {
		return ImageTrainRequest{}
	}
	return ImageTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), LoRARank: int(req.GetLoraRank()),
		Width: int(req.GetWidth()), Height: int(req.GetHeight()), ClassToken: req.GetClassToken(),
	}
}

func ImageTrainRequestToProto(req ImageTrainRequest) *wujiv1.TrainImageRequest {
	return &wujiv1.TrainImageRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, LoraRank: int32(req.LoRARank),
		Width: int32(req.Width), Height: int32(req.Height), ClassToken: req.ClassToken,
	}
}

func VideoTrainRequestFromProto(req *wujiv1.TrainVideoRequest) VideoTrainRequest {
	if req == nil {
		return VideoTrainRequest{}
	}
	return VideoTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), Frames: int(req.GetFrames()),
		FPS: int(req.GetFps()), ContextLength: int(req.GetContextLength()),
	}
}

func VideoTrainRequestToProto(req VideoTrainRequest) *wujiv1.TrainVideoRequest {
	return &wujiv1.TrainVideoRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, Frames: int32(req.Frames),
		Fps: int32(req.FPS), ContextLength: int32(req.ContextLength),
	}
}

func AudioTrainRequestFromProto(req *wujiv1.TrainAudioRequest) AudioTrainRequest {
	if req == nil {
		return AudioTrainRequest{}
	}
	return AudioTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), SampleRate: int(req.GetSampleRate()),
		TaskType: AudioTaskType(req.GetTaskType()),
	}
}

func AudioTrainRequestToProto(req AudioTrainRequest) *wujiv1.TrainAudioRequest {
	return &wujiv1.TrainAudioRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, SampleRate: int32(req.SampleRate),
		TaskType: string(req.TaskType),
	}
}

func Asset3DTrainRequestFromProto(req *wujiv1.Train3DRequest) Asset3DTrainRequest {
	if req == nil {
		return Asset3DTrainRequest{}
	}
	return Asset3DTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), Format: req.GetFormat(),
	}
}

func Asset3DTrainRequestToProto(req Asset3DTrainRequest) *wujiv1.Train3DRequest {
	return &wujiv1.Train3DRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, Format: req.Format,
	}
}

func TTSTrainRequestFromProto(req *wujiv1.TrainTTSRequest) TTSTrainRequest {
	if req == nil {
		return TTSTrainRequest{}
	}
	return TTSTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), Language: req.GetLanguage(),
		ReferenceSample: req.GetReferenceSample(),
	}
}

func TTSTrainRequestToProto(req TTSTrainRequest) *wujiv1.TrainTTSRequest {
	return &wujiv1.TrainTTSRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, Language: req.Language,
		ReferenceSample: req.ReferenceSample,
	}
}

func STTTrainRequestFromProto(req *wujiv1.TrainSTTRequest) STTTrainRequest {
	if req == nil {
		return STTTrainRequest{}
	}
	return STTTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), BaseModel: req.GetBaseModel(),
		OutputPath: req.GetOutputPath(), Epochs: int(req.GetEpochs()),
		LearningRate: req.GetLearningRate(), Language: req.GetLanguage(),
		ModelSize: req.GetModelSize(), FreezeEncoder: req.GetFreezeEncoder(),
	}
}

func STTTrainRequestToProto(req STTTrainRequest) *wujiv1.TrainSTTRequest {
	return &wujiv1.TrainSTTRequest{
		Name: req.Name, DatasetId: req.DatasetID, BaseModel: req.BaseModel,
		OutputPath: req.OutputPath, Epochs: int32(req.Epochs),
		LearningRate: req.LearningRate, Language: req.Language,
		ModelSize: req.ModelSize, FreezeEncoder: req.FreezeEncoder,
	}
}

func VoiceTrainRequestFromProto(req *wujiv1.TrainVoiceRequest) VoiceTrainRequest {
	if req == nil {
		return VoiceTrainRequest{}
	}
	return VoiceTrainRequest{
		Name: req.GetName(), DatasetID: req.GetDatasetId(), OutputPath: req.GetOutputPath(),
		Epochs: int(req.GetEpochs()), SamplePath: req.GetSamplePath(),
		PretrainedModel: req.GetPretrainedModel(), PitchShift: int(req.GetPitchShift()),
		IndexRate: req.GetIndexRate(), BatchSize: int(req.GetBatchSize()),
		SaveEveryEpoch: int(req.GetSaveEveryEpoch()),
	}
}

func VoiceTrainRequestToProto(req VoiceTrainRequest) *wujiv1.TrainVoiceRequest {
	return &wujiv1.TrainVoiceRequest{
		Name: req.Name, DatasetId: req.DatasetID, OutputPath: req.OutputPath,
		Epochs: int32(req.Epochs), SamplePath: req.SamplePath,
		PretrainedModel: req.PretrainedModel, PitchShift: int32(req.PitchShift),
		IndexRate: req.IndexRate, BatchSize: int32(req.BatchSize),
		SaveEveryEpoch: int32(req.SaveEveryEpoch),
	}
}

func TrainResponseFromProto(resp *wujiv1.TrainResponse) *TrainResponse {
	if resp == nil {
		return &TrainResponse{}
	}
	return &TrainResponse{
		JobID: resp.GetJobId(), Status: resp.GetStatus(),
		Capability: capability.Type(resp.GetCapability()), OutputPath: resp.GetOutputPath(),
	}
}

func TrainResponseToProto(resp *TrainResponse) *wujiv1.TrainResponse {
	if resp == nil {
		return &wujiv1.TrainResponse{}
	}
	return &wujiv1.TrainResponse{
		JobId: resp.JobID, Status: resp.Status,
		Capability: string(resp.Capability), OutputPath: resp.OutputPath,
	}
}
