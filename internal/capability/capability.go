package capability

// Type identifies a capability that a driver can provide.
type Type string

const (
	TextGeneration  Type = "text_generation"
	ImageGeneration Type = "image_generation"
	VideoGeneration Type = "video_generation"
	AudioGeneration Type = "audio_generation"
	Asset3D         Type = "3d_asset_generation"
	TTS             Type = "tts"
	STT             Type = "stt"
	VoiceCloning    Type = "voice_cloning"
	Training        Type = "training"
	DatasetMgmt     Type = "dataset_management"
)

// All returns every known capability type.
func All() []Type {
	return []Type{
		TextGeneration,
		ImageGeneration,
		VideoGeneration,
		AudioGeneration,
		Asset3D,
		TTS,
		STT,
		VoiceCloning,
		Training,
		DatasetMgmt,
	}
}

func (t Type) String() string {
	return string(t)
}
