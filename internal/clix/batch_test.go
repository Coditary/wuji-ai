package clix

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildImageRequestBatchHeuristic(t *testing.T) {
	req, err := BuildImageRequest("sunset", false, ImageRequestFields{
		BatchTotal:      12,
		AvailableVRAMMB: 12000,
		Width:           512,
		Height:          512,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.BatchSize != 2 || req.BatchCount != 6 {
		t.Fatalf("batch = %dx%d, want 2x6", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestBatchSizeOnly(t *testing.T) {
	req, err := BuildImageRequest("sunset", false, ImageRequestFields{
		BatchTotal:   12,
		BatchSize:    4,
		BatchSizeSet: true,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.BatchSize != 4 || req.BatchCount != 3 {
		t.Fatalf("batch = %dx%d, want 4x3", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestBatchBothExplicit(t *testing.T) {
	req, err := BuildImageRequest("sunset", false, ImageRequestFields{
		BatchTotal:    12,
		BatchSize:     2,
		BatchCount:    6,
		BatchSizeSet:  true,
		BatchCountSet: true,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.BatchSize != 2 || req.BatchCount != 6 {
		t.Fatalf("batch = %dx%d, want 2x6", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestBatchOnVariation(t *testing.T) {
	req, err := BuildImageRequest("", false, ImageRequestFields{
		InitImage:       "photo.png",
		BatchTotal:      4,
		AvailableVRAMMB: 24000,
		Width:           512,
		Height:          512,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskVariation {
		t.Fatalf("task = %q, want variation", req.Task)
	}
	if req.BatchSize != 4 || req.BatchCount != 1 {
		t.Fatalf("batch = %dx%d, want 4x1", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestVariationWithControlNet(t *testing.T) {
	req, err := BuildImageRequest("", false, ImageRequestFields{
		InitImage: "photo.png",
		ControlUnits: []driver.ControlNetUnit{{
			Type:      driver.ImageControlPose,
			ImagePath: "pose.png",
		}},
		BatchTotal:      4,
		AvailableVRAMMB: 12000,
		Width:           512,
		Height:          512,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskVariation {
		t.Fatalf("task = %q, want variation", req.Task)
	}
	if len(req.ControlUnits) != 1 {
		t.Fatalf("control units = %d, want 1", len(req.ControlUnits))
	}
	if req.BatchSize != 2 || req.BatchCount != 2 {
		t.Fatalf("batch = %dx%d, want 2x2", req.BatchSize, req.BatchCount)
	}

	varReq := req.ToVariationRequest()
	if varReq.InitImagePath != "photo.png" {
		t.Fatalf("init = %q", varReq.InitImagePath)
	}
	if len(varReq.ControlUnits) != 1 {
		t.Fatalf("variation control units = %d, want 1", len(varReq.ControlUnits))
	}
}

func TestBuildImageRequestNoBatchFlag(t *testing.T) {
	req, err := BuildImageRequest("sunset", false, ImageRequestFields{
		BatchSize:  2,
		BatchCount: 3,
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.BatchSize != 2 || req.BatchCount != 3 {
		t.Fatalf("batch = %dx%d, want 2x3", req.BatchSize, req.BatchCount)
	}
}

func TestTaskSupportsBatch(t *testing.T) {
	if !TaskSupportsBatch(driver.ImageTaskGenerate) {
		t.Fatal("generate should support batch")
	}
	if !TaskSupportsBatch(driver.ImageTaskVariation) {
		t.Fatal("variation should support batch")
	}
	if TaskSupportsBatch(driver.ImageTaskUpscale) {
		t.Fatal("upscale should not support batch")
	}
}

func TestResolveImageBatchUpscaleRejected(t *testing.T) {
	_, _, err := ResolveImageBatch(ImageBatchOptions{Total: 4}, driver.ImageTaskUpscale)
	if err == nil {
		t.Fatal("expected error for batch on upscale")
	}
}
