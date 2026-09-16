package cli

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildImageRequestInfersTask(t *testing.T) {
	req, err := buildImageRequest("boat", false, imageRequestFields{
		initImage: "photo.png",
		maskImage: "mask.png",
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskInpaint {
		t.Fatalf("task = %q, want inpaint", req.Task)
	}
}

func TestBuildImageRequestGenerateFromPrompt(t *testing.T) {
	req, err := buildImageRequest("sunset", false, imageRequestFields{})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskGenerate {
		t.Fatalf("task = %q, want generate", req.Task)
	}
}

func TestBuildImageRequestImg2ImgFromInitImage(t *testing.T) {
	req, err := buildImageRequest("watercolor", false, imageRequestFields{
		initImage: "photo.png",
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskImg2Img {
		t.Fatalf("task = %q, want img2img", req.Task)
	}
}

func TestBuildImageRequestEditFromFlag(t *testing.T) {
	req, err := buildImageRequest("make it winter", false, imageRequestFields{
		initImage:     "photo.png",
		editRequested: true,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskEdit {
		t.Fatalf("task = %q, want edit", req.Task)
	}
}

func TestBuildImageRequestUpscaleFromFlag(t *testing.T) {
	req, err := buildImageRequest("optional prompt", false, imageRequestFields{
		initImage:        "photo.png",
		scale:            4,
		upscaleRequested: true,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskUpscale {
		t.Fatalf("task = %q, want upscale", req.Task)
	}
	if req.Scale != 4 {
		t.Fatalf("scale = %g, want 4", req.Scale)
	}
}
