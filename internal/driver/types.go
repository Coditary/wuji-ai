package driver

// TextRequest is the input for text generation.
type TextRequest struct {
	Prompt       string
	Model        string
	SystemPrompt string
	MaxTokens    int
	Temperature  float32

	// Sampling (zero values = use backend default)
	TopP  float32
	TopK  int
	MinP  float32

	// Penalties (zero values = use backend default)
	FrequencyPenalty  float32
	PresencePenalty   float32
	RepetitionPenalty float32

	// Limits
	StopSequences []string

	// Advanced (Seed nil = random; ContextWindow 0 = backend default)
	Seed          *int
	ContextWindow int
}

// TextResponse is the output of text generation.
type TextResponse struct {
	Text         string
	TokensUsed   int
	FinishReason string
}

// ImageRequest is the input for image generation.
type ImageRequest struct {
	Prompt         string
	NegativePrompt string
	Model          string
	Width          int
	Height         int
	Steps          int
	Sampler        string
	CFGScale       float32
	BatchSize      int
	BatchCount     int

	// Advanced (Seed nil = random; zero values = backend default)
	Seed              *int
	DenoisingStrength float32
	InitImagePath     string
	LoRAs             []string
}

// ImageResponse is the output of image generation.
type ImageResponse struct {
	Path   string
	Paths  []string
	Format string
}

// VideoMode identifies the video generation input mode.
type VideoMode string

const (
	VideoModeTextToVideo  VideoMode = "t2v"
	VideoModeImageToVideo VideoMode = "i2v"
)

// CameraControl identifies a preset camera movement.
type CameraControl string

const (
	CameraControlNone     CameraControl = ""
	CameraControlZoomIn   CameraControl = "zoom_in"
	CameraControlZoomOut  CameraControl = "zoom_out"
	CameraControlPanLeft  CameraControl = "pan_left"
	CameraControlPanRight CameraControl = "pan_right"
	CameraControlPanUp    CameraControl = "pan_up"
	CameraControlPanDown  CameraControl = "pan_down"
)

// VideoRequest is the input for video generation.
type VideoRequest struct {
	Prompt   string
	Duration float32
	FPS      int
	Frames   int

	// Motion & dynamics (zero values = backend default)
	MotionStrength float32
	ContextLength  int

	// Methods & quality
	Sampler               string
	Scheduler             string
	Interpolate           bool
	InterpolationStrength float32

	// Input modes
	Mode          VideoMode
	InitImagePath string
	CameraControl CameraControl

	// Optional extras for backends
	NegativePrompt string
	Model          string
	Seed           *int
}

// EffectiveDuration returns video length in seconds from frames/fps or duration.
func (r VideoRequest) EffectiveDuration() float32 {
	if r.Frames > 0 && r.FPS > 0 {
		return float32(r.Frames) / float32(r.FPS)
	}
	return r.Duration
}

// VideoResponse is the output of video generation.
type VideoResponse struct {
	Path     string
	Duration float32
	Frames   int
	FPS      int
}

// AudioTaskType identifies the audio generation mode.
type AudioTaskType string

const (
	AudioTaskTextToMusic   AudioTaskType = "text-to-music"
	AudioTaskMelodyToMusic AudioTaskType = "melody-to-music"
	AudioTaskTextToSFX     AudioTaskType = "text-to-sfx"
	AudioTaskTTS           AudioTaskType = "tts"
)

// AudioRequest is the input for audio generation.
type AudioRequest struct {
	Prompt         string
	Lyrics         string
	NegativePrompt string
	Model          string
	Duration       float32

	// Structure (zero values = backend default)
	Overlap float32

	// Sampling (zero values = backend default)
	Temperature float32
	CFGScale    float32
	TopP        float32
	TopK        int

	// Format & technical
	SampleRate    int
	TaskType      AudioTaskType
	ReferencePath string
	Seed          *int
}

// AudioResponse is the output of audio generation.
type AudioResponse struct {
	Path       string
	Duration   float32
	SampleRate int
	Format     string
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
	Text              string
	Voice             string
	VoiceSamplePath   string
	Language          string
	Speed             float32
	Emotion           string
	Temperature       float32
	RepetitionPenalty float32
	Model             string
	Seed              *int
}

// TTSResponse is the output of text-to-speech.
type TTSResponse struct {
	Path       string
	Duration   float32
	SampleRate int
	Format     string
}

// STTTask identifies the speech-to-text operation mode.
type STTTask string

const (
	STTTaskTranscribe STTTask = "transcribe"
	STTTaskTranslate  STTTask = "translate"
)

// STTRequest is the input for speech-to-text.
type STTRequest struct {
	AudioPath      string
	Language       string
	Model          string
	Task           STTTask
	BeamSize       int
	WordTimestamps bool
	VADEnabled     bool
	VADThreshold   float32
}

// WordTimestamp marks when a word was spoken in the transcript.
type WordTimestamp struct {
	Word  string
	Start float32
	End   float32
}

// STTResponse is the output of speech-to-text.
type STTResponse struct {
	Text       string
	Confidence float32
	Words      []WordTimestamp
}

// VoiceCloneMode identifies the voice cloning workflow.
type VoiceCloneMode string

const (
	VoiceCloneModeZeroShot   VoiceCloneMode = "zero-shot"
	VoiceCloneModeConversion VoiceCloneMode = "conversion"
)

// VoiceRequest is the input for voice cloning.
type VoiceRequest struct {
	Name         string
	SamplePath   string
	TargetModel  string
	SourcePath   string
	Mode         VoiceCloneMode
	PitchShift   int
	IndexRate    float32
	Protect      float32
	Vocoder      string
	Denoise      bool
	DenoiseStrength float32
	ChunkSize    int
	Crossfade    float32
}

// VoiceResponse is the output of voice cloning.
type VoiceResponse struct {
	VoiceID    string
	Name       string
	OutputPath string
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
