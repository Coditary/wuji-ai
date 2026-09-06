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

// TextTrainer fine-tunes text generation models.
type TextTrainer interface {
	TrainText(ctx context.Context, req TextTrainRequest) (*TrainResponse, error)
}

// ImageTrainer fine-tunes image generation models.
type ImageTrainer interface {
	TrainImage(ctx context.Context, req ImageTrainRequest) (*TrainResponse, error)
}

// VideoTrainer fine-tunes video generation models.
type VideoTrainer interface {
	TrainVideo(ctx context.Context, req VideoTrainRequest) (*TrainResponse, error)
}

// AudioTrainer fine-tunes audio generation models.
type AudioTrainer interface {
	TrainAudio(ctx context.Context, req AudioTrainRequest) (*TrainResponse, error)
}

// Asset3DTrainer fine-tunes 3D asset models.
type Asset3DTrainer interface {
	Train3D(ctx context.Context, req Asset3DTrainRequest) (*TrainResponse, error)
}

// TTSTrainer fine-tunes text-to-speech models.
type TTSTrainer interface {
	TrainTTS(ctx context.Context, req TTSTrainRequest) (*TrainResponse, error)
}

// STTTrainer fine-tunes speech-to-text models.
type STTTrainer interface {
	TrainSTT(ctx context.Context, req STTTrainRequest) (*TrainResponse, error)
}

// VoiceTrainer fine-tunes voice cloning models.
type VoiceTrainer interface {
	TrainVoice(ctx context.Context, req VoiceTrainRequest) (*TrainResponse, error)
}

// DatasetManager manages datasets.
type DatasetManager interface {
	ManageDataset(ctx context.Context, req DatasetRequest) (*DatasetResponse, error)
}
