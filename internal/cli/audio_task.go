package cli

import (
	"github.com/coditary/wuji-core/pkg/driver"
)

func buildAudioRequest(prompt string, fields audioRequestFields) (driver.AudioRequest, error) {
	task, err := driver.InferAudioTask(driver.AudioTaskInputs{
		ReferencePath:   fields.referencePath,
		SFXRequested:    fields.sfxRequested,
		SpeechRequested: fields.speechRequested,
		VoiceName:       fields.voice,
	})
	if err != nil {
		return driver.AudioRequest{}, err
	}

	duration := fields.duration
	if task == driver.AudioTaskTextToSpeech && !fields.durationSet {
		duration = 0
	}

	req := driver.AudioRequest{
		Task: task, Prompt: prompt, Lyrics: fields.lyrics, NegativePrompt: fields.negativePrompt,
		Model: fields.model, Duration: duration, Overlap: fields.overlap, Temperature: fields.temperature,
		CFGScale: fields.cfgScale, TopP: fields.topP, TopK: fields.topK, SampleRate: fields.sampleRate,
		ReferencePath: fields.referencePath, Voice: fields.voice, Language: fields.language,
		Speed: fields.speed, Pitch: fields.pitch, Emotion: fields.emotion, Style: fields.style,
		Energy: fields.energy, Format: fields.format,
	}
	if fields.seed >= 0 {
		req.Seed = &fields.seed
	}
	if err := req.Validate(); err != nil {
		return driver.AudioRequest{}, err
	}
	return req, nil
}

type audioRequestFields struct {
	lyrics          string
	negativePrompt  string
	model           string
	duration        float32
	durationSet     bool
	overlap         float32
	temperature     float32
	cfgScale        float32
	topP            float32
	topK            int
	sampleRate      int
	referencePath   string
	voice           string
	language        string
	speed           float32
	pitch           int
	emotion         string
	style           string
	energy          float32
	format          string
	sfxRequested    bool
	speechRequested bool
	seed            int
}
