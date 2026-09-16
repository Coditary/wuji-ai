package cli

import (
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildAudioRequestTextToMusic(t *testing.T) {
	req, err := buildAudioRequest("lofi beat", audioRequestFields{})
	if err != nil {
		t.Fatalf("buildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToMusic {
		t.Fatalf("task = %q, want text-to-music", req.Task)
	}
}

func TestBuildAudioRequestMelodyFromReference(t *testing.T) {
	req, err := buildAudioRequest("epic theme", audioRequestFields{
		referencePath: "melody.mid",
	})
	if err != nil {
		t.Fatalf("buildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskMelodyToMusic || req.ReferencePath != "melody.mid" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildAudioRequestSFXFromFlag(t *testing.T) {
	req, err := buildAudioRequest("thunder", audioRequestFields{
		sfxRequested: true,
	})
	if err != nil {
		t.Fatalf("buildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToSFX {
		t.Fatalf("task = %q, want text-to-sfx", req.Task)
	}
}

func TestBuildAudioRequestReferenceAndSFXConflict(t *testing.T) {
	_, err := buildAudioRequest("conflict", audioRequestFields{
		referencePath: "melody.mid",
		sfxRequested:  true,
	})
	if err == nil || !strings.Contains(err.Error(), "--reference") {
		t.Fatalf("expected reference/sfx conflict, got %v", err)
	}
}

func TestBuildAudioRequestSpeechFromVoice(t *testing.T) {
	req, err := buildAudioRequest("Hallo Welt", audioRequestFields{
		voice: "myvoice", language: "de", speed: 1.1, pitch: 1,
		emotion: "happy", style: "conversational", energy: 0.7,
	})
	if err != nil {
		t.Fatalf("buildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToSpeech || req.Voice != "myvoice" || req.Language != "de" ||
		req.Emotion != "happy" || req.Style != "conversational" || req.Energy != 0.7 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildAudioRequestSpeechFromFlag(t *testing.T) {
	req, err := buildAudioRequest("Hello", audioRequestFields{speechRequested: true})
	if err != nil {
		t.Fatalf("buildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToSpeech {
		t.Fatalf("task = %q, want text-to-speech", req.Task)
	}
}

func TestBuildAudioRequestSpeechFieldsOnMusicRejected(t *testing.T) {
	_, err := buildAudioRequest("beat", audioRequestFields{language: "de"})
	if err == nil || !strings.Contains(err.Error(), "text-to-speech") {
		t.Fatalf("expected speech-only flag error, got %v", err)
	}
}
