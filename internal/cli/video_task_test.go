package cli

import (
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildVideoRequestGenerate(t *testing.T) {
	req, err := buildVideoRequest("waves", false, videoRequestFields{})
	if err != nil {
		t.Fatalf("buildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskGenerate {
		t.Fatalf("task = %q", req.Task)
	}
}

func TestBuildVideoRequestImageToVideo(t *testing.T) {
	req, err := buildVideoRequest("animate", false, videoRequestFields{
		imagePath: "frame.png",
	})
	if err != nil {
		t.Fatalf("buildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskImageToVideo || req.InitImagePath != "frame.png" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildVideoRequestInterpolateFromFPS(t *testing.T) {
	req, err := buildVideoRequest("", false, videoRequestFields{
		videoPath: "clip.mp4",
		fps:       60,
		fpsSet:    true,
	})
	if err != nil {
		t.Fatalf("buildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskInterpolate || req.FPS != 60 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildVideoRequestCameraOnlyOnGenerate(t *testing.T) {
	_, err := buildVideoRequest("animate", false, videoRequestFields{
		imagePath: "frame.png", cameraControl: "zoom_in",
	})
	if err == nil || !strings.Contains(err.Error(), "--camera") {
		t.Fatalf("expected camera error, got %v", err)
	}
}

func TestBuildVideoRequestScaleFromScale(t *testing.T) {
	req, err := buildVideoRequest("", false, videoRequestFields{
		videoPath: "clip.mp4",
		scale:     4,
		scaleSet:  true,
	})
	if err != nil {
		t.Fatalf("buildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskScale || req.Scale != 4 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildVideoRequestScaleFractional(t *testing.T) {
	req, err := buildVideoRequest("", false, videoRequestFields{
		videoPath: "clip.mp4",
		scale:     2.5,
		scaleSet:  true,
	})
	if err != nil {
		t.Fatalf("buildVideoRequest: %v", err)
	}
	if req.TaskOrDefault() != driver.VideoTaskScale || req.Scale != 2.5 {
		t.Fatalf("unexpected request: %+v", req)
	}
}
