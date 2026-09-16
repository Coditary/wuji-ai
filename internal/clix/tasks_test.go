package clix

import (
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestParseImageScaleInput(t *testing.T) {
	v, requested, err := ParseImageScaleInput("250%", false)
	if err != nil {
		t.Fatal(err)
	}
	if v != 2.5 || !requested {
		t.Fatalf("got scale=%g requested=%v", v, requested)
	}

	v, requested, err = ParseImageScaleInput("", true)
	if err != nil {
		t.Fatal(err)
	}
	if v != 4 || !requested {
		t.Fatalf("default upscale: got scale=%g requested=%v", v, requested)
	}
}

func TestBuildImageRequestInfersTask(t *testing.T) {
	req, err := BuildImageRequest("boat", false, ImageRequestFields{
		InitImage: "photo.png",
		MaskImage: "mask.png",
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskInpaint {
		t.Fatalf("task = %q, want inpaint", req.Task)
	}
}

func TestBuildImageRequestGenerateFromPrompt(t *testing.T) {
	req, err := BuildImageRequest("sunset", false, ImageRequestFields{})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskGenerate {
		t.Fatalf("task = %q, want generate", req.Task)
	}
}

func TestBuildImageRequestImg2ImgFromInitImage(t *testing.T) {
	req, err := BuildImageRequest("watercolor", false, ImageRequestFields{
		InitImage: "photo.png",
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskImg2Img {
		t.Fatalf("task = %q, want img2img", req.Task)
	}
}

func TestBuildImageRequestEditFromFlag(t *testing.T) {
	req, err := BuildImageRequest("make it winter", false, ImageRequestFields{
		InitImage:     "photo.png",
		EditRequested: true,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskEdit {
		t.Fatalf("task = %q, want edit", req.Task)
	}
}

func TestBuildImageRequestUpscaleFromFlag(t *testing.T) {
	req, err := BuildImageRequest("optional prompt", false, ImageRequestFields{
		InitImage:        "photo.png",
		Scale:            4,
		UpscaleRequested: true,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskUpscale {
		t.Fatalf("task = %q, want upscale", req.Task)
	}
	if req.Scale != 4 {
		t.Fatalf("scale = %g, want 4", req.Scale)
	}
}

func TestBuildAudioRequestTextToMusic(t *testing.T) {
	req, err := BuildAudioRequest("lofi beat", AudioRequestFields{})
	if err != nil {
		t.Fatalf("BuildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToMusic {
		t.Fatalf("task = %q, want text-to-music", req.Task)
	}
}

func TestBuildAudioRequestMelodyFromReference(t *testing.T) {
	req, err := BuildAudioRequest("epic theme", AudioRequestFields{
		ReferencePath: "melody.mid",
	})
	if err != nil {
		t.Fatalf("BuildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskMelodyToMusic || req.ReferencePath != "melody.mid" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildAudioRequestSFXFromFlag(t *testing.T) {
	req, err := BuildAudioRequest("thunder", AudioRequestFields{
		SFXRequested: true,
	})
	if err != nil {
		t.Fatalf("BuildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToSFX {
		t.Fatalf("task = %q, want text-to-sfx", req.Task)
	}
}

func TestBuildAudioRequestReferenceAndSFXConflict(t *testing.T) {
	_, err := BuildAudioRequest("conflict", AudioRequestFields{
		ReferencePath: "melody.mid",
		SFXRequested:  true,
	})
	if err == nil || !strings.Contains(err.Error(), "--reference") {
		t.Fatalf("expected reference/sfx conflict, got %v", err)
	}
}

func TestBuildAudioRequestSpeechFromVoice(t *testing.T) {
	req, err := BuildAudioRequest("Hallo Welt", AudioRequestFields{
		Voice: "myvoice", Language: "de", Speed: 1.1, Pitch: 1,
		Emotion: "happy", Style: "conversational", Energy: 0.7,
	})
	if err != nil {
		t.Fatalf("BuildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToSpeech || req.Voice != "myvoice" || req.Language != "de" ||
		req.Emotion != "happy" || req.Style != "conversational" || req.Energy != 0.7 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildAudioRequestSpeechFromFlag(t *testing.T) {
	req, err := BuildAudioRequest("Hello", AudioRequestFields{SpeechRequested: true})
	if err != nil {
		t.Fatalf("BuildAudioRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.AudioTaskTextToSpeech {
		t.Fatalf("task = %q, want text-to-speech", req.Task)
	}
}

func TestBuildAudioRequestSpeechFieldsOnMusicRejected(t *testing.T) {
	_, err := BuildAudioRequest("beat", AudioRequestFields{Language: "de"})
	if err == nil || !strings.Contains(err.Error(), "text-to-speech") {
		t.Fatalf("expected speech-only flag error, got %v", err)
	}
}

func TestBuildVideoRequestGenerate(t *testing.T) {
	req, err := BuildVideoRequest("waves", false, VideoRequestFields{})
	if err != nil {
		t.Fatalf("BuildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskGenerate {
		t.Fatalf("task = %q", req.Task)
	}
}

func TestBuildVideoRequestImageToVideo(t *testing.T) {
	req, err := BuildVideoRequest("animate", false, VideoRequestFields{
		ImagePath: "frame.png",
	})
	if err != nil {
		t.Fatalf("BuildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskImageToVideo || req.InitImagePath != "frame.png" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildVideoRequestInterpolateFromFPS(t *testing.T) {
	req, err := BuildVideoRequest("", false, VideoRequestFields{
		VideoPath: "clip.mp4",
		FPS:       60,
		FPSSet:    true,
	})
	if err != nil {
		t.Fatalf("BuildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskInterpolate || req.FPS != 60 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildVideoRequestCameraOnlyOnGenerate(t *testing.T) {
	_, err := BuildVideoRequest("animate", false, VideoRequestFields{
		ImagePath: "frame.png", CameraControl: "zoom_in",
	})
	if err == nil || !strings.Contains(err.Error(), "--camera") {
		t.Fatalf("expected camera error, got %v", err)
	}
}

func TestBuildVideoRequestScaleFromScale(t *testing.T) {
	req, err := BuildVideoRequest("", false, VideoRequestFields{
		VideoPath: "clip.mp4",
		Scale:     4,
		ScaleSet:  true,
	})
	if err != nil {
		t.Fatalf("BuildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskScale || req.Scale != 4 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildMeshRequestGenerate(t *testing.T) {
	req, err := BuildMeshRequest("low poly crate", false, MeshRequestFields{Format: "glb"})
	if err != nil {
		t.Fatalf("BuildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskGenerate {
		t.Fatalf("task = %q, want generate", req.Task)
	}
}

func TestBuildMeshRequestImageToMesh(t *testing.T) {
	req, err := BuildMeshRequest("", false, MeshRequestFields{ImagePath: "ref.png"})
	if err != nil {
		t.Fatalf("BuildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskImageToMesh {
		t.Fatalf("task = %q, want i2m", req.Task)
	}
}

func TestBuildMeshRequestExplicitTask(t *testing.T) {
	req, err := BuildMeshRequest("x", false, MeshRequestFields{
		TaskExplicit: "repair",
		MeshPath:     "scan.glb",
	})
	if err != nil {
		t.Fatalf("BuildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskRepair {
		t.Fatalf("task = %q, want repair", req.Task)
	}
}

func TestBuildMeshRequestScene(t *testing.T) {
	req, err := BuildMeshRequest("medieval room", false, MeshRequestFields{SceneRequested: true})
	if err != nil {
		t.Fatalf("BuildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskScene {
		t.Fatalf("task = %q, want scene", req.Task)
	}
}

func TestBuildMeshRequestMode(t *testing.T) {
	req, err := BuildMeshRequest("knight", false, MeshRequestFields{Mode: "human"})
	if err != nil {
		t.Fatalf("BuildMeshRequest: %v", err)
	}
	if req.Mode != driver.MeshModeHuman {
		t.Fatalf("mode = %q, want human", req.Mode)
	}
}

func TestBuildMeshRequestUpscale(t *testing.T) {
	req, err := BuildMeshRequest("", false, MeshRequestFields{
		MeshPath:         "hero.glb",
		UpscaleRequested: true,
		Scale:            2,
		ScaleSet:         true,
	})
	if err != nil {
		t.Fatalf("BuildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskUpscale {
		t.Fatalf("task = %q, want upscale", req.Task)
	}
}
