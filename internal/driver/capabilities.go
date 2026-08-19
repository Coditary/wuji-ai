package driver

import "context"

// TextGenerator generates text from prompts.
type TextGenerator interface {
	GenerateText(ctx context.Context, req TextRequest) (*TextResponse, error)
}

// ImageGenerator generates images from prompts.
type ImageGenerator interface {
	GenerateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error)
}

// VideoGenerator generates videos from prompts.
type VideoGenerator interface {
	GenerateVideo(ctx context.Context, req VideoRequest) (*VideoResponse, error)
}

// AudioGenerator generates audio from prompts.
type AudioGenerator interface {
	GenerateAudio(ctx context.Context, req AudioRequest) (*AudioResponse, error)
}

// Asset3DGenerator generates 3D assets from prompts.
type Asset3DGenerator interface {
	Generate3D(ctx context.Context, req Asset3DRequest) (*Asset3DResponse, error)
}

// SpeechSynthesizer converts text to speech.
type SpeechSynthesizer interface {
	Synthesize(ctx context.Context, req TTSRequest) (*TTSResponse, error)
}

// SpeechTranscriber converts speech to text.
type SpeechTranscriber interface {
	Transcribe(ctx context.Context, req STTRequest) (*STTResponse, error)
}

// VoiceCloner creates custom voices from samples.
type VoiceCloner interface {
	CloneVoice(ctx context.Context, req VoiceRequest) (*VoiceResponse, error)
}

// Trainer runs model training jobs.
type Trainer interface {
	Train(ctx context.Context, req TrainRequest) (*TrainResponse, error)
}

// DatasetManager manages datasets.
type DatasetManager interface {
	ManageDataset(ctx context.Context, req DatasetRequest) (*DatasetResponse, error)
}
