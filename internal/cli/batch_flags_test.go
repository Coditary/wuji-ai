package cli

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildImageRequestBatchHeuristic(t *testing.T) {
	req, err := buildImageRequest("sunset", false, imageRequestFields{
		batchTotal:      12,
		availableVRAMMB: 12000,
		width:           512,
		height:          512,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.BatchSize != 2 || req.BatchCount != 6 {
		t.Fatalf("batch = %dx%d, want 2x6", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestBatchSizeOnly(t *testing.T) {
	req, err := buildImageRequest("sunset", false, imageRequestFields{
		batchTotal:   12,
		batchSize:    4,
		batchSizeSet: true,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.BatchSize != 4 || req.BatchCount != 3 {
		t.Fatalf("batch = %dx%d, want 4x3", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestBatchBothExplicit(t *testing.T) {
	req, err := buildImageRequest("sunset", false, imageRequestFields{
		batchTotal:    12,
		batchSize:     2,
		batchCount:    6,
		batchSizeSet:  true,
		batchCountSet: true,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.BatchSize != 2 || req.BatchCount != 6 {
		t.Fatalf("batch = %dx%d, want 2x6", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestBatchOnVariation(t *testing.T) {
	req, err := buildImageRequest("", false, imageRequestFields{
		initImage:       "photo.png",
		batchTotal:      4,
		availableVRAMMB: 24000,
		width:           512,
		height:          512,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskVariation {
		t.Fatalf("task = %q, want variation", req.Task)
	}
	if req.BatchSize != 4 || req.BatchCount != 1 {
		t.Fatalf("batch = %dx%d, want 4x1", req.BatchSize, req.BatchCount)
	}
}

func TestBuildImageRequestVariationWithControlNet(t *testing.T) {
	req, err := buildImageRequest("", false, imageRequestFields{
		initImage: "photo.png",
		controlUnits: []driver.ControlNetUnit{{
			Type:      driver.ImageControlPose,
			ImagePath: "pose.png",
		}},
		batchTotal:      4,
		availableVRAMMB: 12000,
		width:           512,
		height:          512,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
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
	req, err := buildImageRequest("sunset", false, imageRequestFields{
		batchSize:  2,
		batchCount: 3,
	})
	if err != nil {
		t.Fatalf("buildImageRequest: %v", err)
	}
	if req.BatchSize != 2 || req.BatchCount != 3 {
		t.Fatalf("batch = %dx%d, want 2x3", req.BatchSize, req.BatchCount)
	}
}

func TestTaskSupportsBatch(t *testing.T) {
	if !taskSupportsBatch(driver.ImageTaskGenerate) {
		t.Fatal("generate should support batch")
	}
	if taskSupportsBatch(driver.ImageTaskVariation) != true {
		t.Fatal("variation should support batch")
	}
	if taskSupportsBatch(driver.ImageTaskUpscale) != false {
		t.Fatal("upscale should not support batch")
	}
}
