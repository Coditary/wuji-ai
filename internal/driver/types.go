package driver

// TextRequest is the input for text generation.
type TextRequest struct {
	Prompt      string
	Model       string
	MaxTokens   int
	Temperature float32
}

// TextResponse is the output of text generation.
type TextResponse struct {
	Text         string
	TokensUsed   int
	FinishReason string
}

// ImageRequest is the input for image generation.
type ImageRequest struct {
	Prompt string
	Width  int
	Height int
	Steps  int
}

// ImageResponse is the output of image generation.
type ImageResponse struct {
	Path   string
	Format string
}

// VideoRequest is the input for video generation.
type VideoRequest struct {
	Prompt   string
	Duration float32
	FPS      int
}

// VideoResponse is the output of video generation.
type VideoResponse struct {
	Path     string
	Duration float32
}

// AudioRequest is the input for audio generation.
type AudioRequest struct {
	Prompt   string
	Duration float32
}

// AudioResponse is the output of audio generation.
type AudioResponse struct {
	Path     string
	Duration float32
}

// Asset3DRequest is the input for 3D asset generation.
type Asset3DRequest struct {
	Prompt string
	Format string
}

// Asset3DResponse is the output of 3D asset generation.
type Asset3DResponse struct {
	Path   string
	Format string
}

// TTSRequest is the input for text-to-speech.
type TTSRequest struct {
	Text  string
	Voice string
}

// TTSResponse is the output of text-to-speech.
type TTSResponse struct {
	Path     string
	Duration float32
}

// STTRequest is the input for speech-to-text.
type STTRequest struct {
	AudioPath string
	Language  string
}

// STTResponse is the output of speech-to-text.
type STTResponse struct {
	Text       string
	Confidence float32
}

// VoiceRequest is the input for voice cloning.
type VoiceRequest struct {
	SamplePath string
	Name       string
}

// VoiceResponse is the output of voice cloning.
type VoiceResponse struct {
	VoiceID string
	Name    string
}

// TrainRequest is the input for model training.
type TrainRequest struct {
	DatasetID string
	ModelType string
	Epochs    int
}

// TrainResponse is the output of model training.
type TrainResponse struct {
	JobID  string
	Status string
}

// DatasetAction identifies a dataset management operation.
type DatasetAction string

const (
	DatasetList   DatasetAction = "list"
	DatasetCreate DatasetAction = "create"
	DatasetDelete DatasetAction = "delete"
)

// DatasetRequest is the input for dataset management.
type DatasetRequest struct {
	Action DatasetAction
	Name   string
	Path   string
}

// DatasetEntry describes a single dataset.
type DatasetEntry struct {
	ID   string
	Name string
	Path string
	Size int64
}

// DatasetResponse is the output of dataset management.
type DatasetResponse struct {
	Datasets []DatasetEntry
	Message  string
}
