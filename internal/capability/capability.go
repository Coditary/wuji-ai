package capability

// Type identifies a capability that a driver can provide.
type Type string

const (
	TextGeneration  Type = "text"
	ImageGeneration Type = "image"
	VideoGeneration Type = "video"
	AudioGeneration Type = "audio"
	Asset3D         Type = "3d"
	TTS             Type = "tts"
	STT             Type = "stt"
	VoiceCloning    Type = "voice"
	Training        Type = "train"
	DatasetMgmt     Type = "dataset"
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
