package driver

import "github.com/coditary/wuji/internal/capability"

// TrainResponse is the output of any training job.
type TrainResponse struct {
	JobID      string
	Status     string
	Capability capability.Type
	OutputPath string
}

// TextTrainRequest configures LLM / text model fine-tuning.
type TextTrainRequest struct {
	Name          string
	DatasetID     string
	BaseModel     string
	OutputPath    string
	Epochs        int
	LearningRate  float32
	LoRARank      int
	ContextLength int
	BatchSize     int
}

// ImageTrainRequest configures image / diffusion model training.
type ImageTrainRequest struct {
	Name         string
	DatasetID    string
	BaseModel    string
	OutputPath   string
	Epochs       int
	LearningRate float32
	LoRARank     int
	Width        int
	Height       int
	ClassToken   string
}

// VideoTrainRequest configures video model training.
type VideoTrainRequest struct {
	Name          string
	DatasetID     string
	BaseModel     string
	OutputPath    string
	Epochs        int
	LearningRate  float32
	Frames        int
	FPS           int
	ContextLength int
}

// AudioTrainRequest configures audio / music model training.
type AudioTrainRequest struct {
	Name         string
	DatasetID    string
	BaseModel    string
	OutputPath   string
	Epochs       int
	LearningRate float32
	SampleRate   int
	TaskType     AudioTaskType
}

// Asset3DTrainRequest configures 3D asset model training.
type Asset3DTrainRequest struct {
	Name         string
	DatasetID    string
	BaseModel    string
	OutputPath   string
	Epochs       int
	LearningRate float32
	Format       string
}

// TTSTrainRequest configures text-to-speech voice / model training.
type TTSTrainRequest struct {
	Name            string
	DatasetID       string
	BaseModel       string
	OutputPath      string
	Epochs          int
	LearningRate    float32
	Language        string
	ReferenceSample string
}

// STTTrainRequest configures speech-to-text model fine-tuning.
type STTTrainRequest struct {
	Name          string
	DatasetID     string
	BaseModel     string
	OutputPath    string
	Epochs        int
	LearningRate  float32
	Language      string
	ModelSize     string
	FreezeEncoder bool
}

// VoiceTrainRequest configures voice cloning model training (e.g. RVC).
type VoiceTrainRequest struct {
	Name            string
	DatasetID       string
	OutputPath      string
	Epochs          int
	SamplePath      string
	PretrainedModel string
	PitchShift      int
	IndexRate       float32
	BatchSize       int
	SaveEveryEpoch  int
}
