package clix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildTextRequestFromPrompt(t *testing.T) {
	req, err := BuildTextRequest([]string{"hello"}, false, TextRequestFields{Model: "m"})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if req.InputModeOrDefault() != driver.TextInputPrompt || req.Prompt != "hello" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildTextRequestFromTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(path, []byte("from file"), 0o644); err != nil {
		t.Fatal(err)
	}
	req, err := BuildTextRequest(nil, false, TextRequestFields{TextFile: path})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if req.Prompt != "from file" {
		t.Fatalf("prompt = %q", req.Prompt)
	}
}

func TestBuildTextRequestFromMessagesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.json")
	if err := os.WriteFile(path, []byte(`[
		{"role":"system","content":"sys"},
		{"role":"user","content":"hi"}
	]`), 0o644); err != nil {
		t.Fatal(err)
	}
	req, err := BuildTextRequest(nil, false, TextRequestFields{MessagesFile: path, Model: "m"})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if len(req.Messages) != 2 || req.Messages[1].Content != "hi" {
		t.Fatalf("unexpected messages: %+v", req.Messages)
	}
	if req.Prompt != "" {
		t.Fatalf("expected empty prompt, got %q", req.Prompt)
	}
}

func TestBuildTextRequestMutuallyExclusiveMedia(t *testing.T) {
	_, err := BuildTextRequest(nil, false, TextRequestFields{
		ImagePath: "a.jpg",
		AudioPath: "b.wav",
	})
	if err == nil {
		t.Fatal("expected error for multiple media flags")
	}
}

func TestBuildTextRequestPromptAndTextFile(t *testing.T) {
	_, err := BuildTextRequest([]string{"hi"}, false, TextRequestFields{TextFile: "x.txt"})
	if err == nil {
		t.Fatal("expected error for prompt and --text together")
	}
}

func TestBuildTextRequestAudioMode(t *testing.T) {
	req, err := BuildTextRequest(nil, false, TextRequestFields{AudioPath: "meeting.mp4"})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if req.InputModeOrDefault() != driver.TextInputAudio || req.MediaPath != "meeting.mp4" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildTextRequestTranslateRequiresText(t *testing.T) {
	_, err := BuildTextRequest(nil, false, TextRequestFields{
		AudioPath: "a.wav",
		Translate: true,
	})
	if err == nil || !strings.Contains(err.Error(), "--translate") {
		t.Fatalf("expected translate error, got %v", err)
	}
}

func TestBuildTextRequestAudioFlagsRequireAudio(t *testing.T) {
	_, err := BuildTextRequest([]string{"hi"}, false, TextRequestFields{Quality: "careful"})
	if err == nil || !strings.Contains(err.Error(), "--audio") {
		t.Fatalf("expected audio flag error, got %v", err)
	}
}

func TestBuildTextRequestFromStdinFlag(t *testing.T) {
	req, err := BuildTextRequest(nil, false, TextRequestFields{
		UseStdin:    true,
		StdinReader: strings.NewReader("from stdin"),
	})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if req.Prompt != "from stdin" {
		t.Fatalf("prompt = %q", req.Prompt)
	}
}

func TestBuildTextRequestFromPipedStdin(t *testing.T) {
	req, err := BuildTextRequest(nil, false, TextRequestFields{
		StdinReader: strings.NewReader("piped in"),
	})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if req.Prompt != "piped in" {
		t.Fatalf("prompt = %q", req.Prompt)
	}
}

func TestBuildTextRequestPromptAndStdin(t *testing.T) {
	_, err := BuildTextRequest([]string{"hi"}, false, TextRequestFields{UseStdin: true})
	if err == nil {
		t.Fatal("expected error for prompt and --stdin together")
	}
}

func TestBuildTextRequestAudioWithInferenceFlags(t *testing.T) {
	req, err := BuildTextRequest(nil, false, TextRequestFields{
		AudioPath: "a.wav", Language: "de", Quality: "balanced", TrimSilence: true, SilenceLevel: 0.4,
	})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if req.Language != "de" || req.BeamSize != 3 || !req.VADEnabled || req.VADThreshold != 0.4 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildTextRequestTrainMode(t *testing.T) {
	req, err := BuildTextRequest([]string{"train", "me"}, true, TextRequestFields{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if req.Prompt != "train me" || req.Model != "m" {
		t.Fatalf("req=%+v", req)
	}
}

func TestParseAudioQualityPresets(t *testing.T) {
	cases := map[string]int{
		"fast": 1, "balanced": 3, "careful": 5, "4": 4,
	}
	for input, want := range cases {
		got, err := ParseAudioQuality(input)
		if err != nil {
			t.Fatalf("ParseAudioQuality(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseAudioQuality(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestParseAudioQualityInvalid(t *testing.T) {
	for _, input := range []string{"turbo", "0", "9"} {
		if _, err := ParseAudioQuality(input); err == nil {
			t.Fatalf("expected error for quality %q", input)
		}
	}
}

func TestSilenceLevelEnablesTrimSilence(t *testing.T) {
	req, err := BuildTextRequest(nil, false, TextRequestFields{
		AudioPath: "a.wav", SilenceLevel: 0.2,
	})
	if err != nil {
		t.Fatalf("BuildTextRequest: %v", err)
	}
	if !req.VADEnabled {
		t.Fatal("expected trim silence when silence-level is set")
	}
}

func TestValidateSilenceLevel(t *testing.T) {
	if err := ValidateSilenceLevel(1.5); err == nil {
		t.Fatal("expected error for silence-level > 1")
	}
}
