package clix

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/spf13/cobra"
)

func TestImageControlFlagsBuildUnits(t *testing.T) {
	cmd := &cobra.Command{Use: "image"}
	opts := &ImageControlOptions{}
	RegisterImageControlFlags(cmd, opts)

	if err := cmd.ParseFlags([]string{
		"--pose=full", "--pose", "pose.png", "--pose-weight", "0.9",
		"--depth=midas", "--depth", "photo.jpg",
		"--edge", "edges.png", "--canny-weight", "0.7",
		"--control-mode", "balanced",
	}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	units, err := opts.BuildUnits(cmd)
	if err != nil {
		t.Fatalf("BuildUnits: %v", err)
	}
	if len(units) != 3 {
		t.Fatalf("len(units) = %d, want 3", len(units))
	}

	mode, err := opts.ControlMode()
	if err != nil {
		t.Fatalf("ControlMode: %v", err)
	}
	if mode != driver.ControlNetModeBalanced {
		t.Fatalf("mode = %q", mode)
	}
}

func TestImageControlSynonymParamFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "image"}
	opts := &ImageControlOptions{}
	RegisterImageControlFlags(cmd, opts)

	if err := cmd.ParseFlags([]string{
		"--edge", "edges.png",
		"--edge-weight", "0.85",
		"--edge-low", "90",
		"--edge-high", "210",
		"--openpose", "pose.png",
		"--openpose-weight", "0.75",
		"--openpose=full",
	}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	units, err := opts.BuildUnits(cmd)
	if err != nil {
		t.Fatalf("BuildUnits: %v", err)
	}
	if len(units) != 2 {
		t.Fatalf("len(units) = %d, want 2", len(units))
	}

	var canny, pose *driver.ControlNetUnit
	for i := range units {
		switch units[i].Type {
		case driver.ImageControlCanny:
			canny = &units[i]
		case driver.ImageControlPose:
			pose = &units[i]
		}
	}
	if canny == nil || pose == nil {
		t.Fatalf("units = %+v", units)
	}
	if canny.Weight != 0.85 {
		t.Fatalf("canny weight = %g, want 0.85", canny.Weight)
	}
	if canny.ThresholdA != 90 || canny.ThresholdB != 210 {
		t.Fatalf("canny thresholds = %g/%g, want 90/210", canny.ThresholdA, canny.ThresholdB)
	}
	if pose.Weight != 0.75 || pose.Preprocessor != "full" {
		t.Fatalf("pose = %+v", pose)
	}
}

func TestImageControlCannyAndEdgeTogether(t *testing.T) {
	cmd := &cobra.Command{Use: "image"}
	opts := &ImageControlOptions{}
	RegisterImageControlFlags(cmd, opts)

	if err := cmd.ParseFlags([]string{
		"--canny", "first.png",
		"--edge", "final.png",
		"--edge-weight", "0.6",
	}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	units, err := opts.BuildUnits(cmd)
	if err != nil {
		t.Fatalf("BuildUnits: %v", err)
	}
	if len(units) != 1 {
		t.Fatalf("len(units) = %d, want 1 (same canonical type)", len(units))
	}
	if units[0].Type != driver.ImageControlCanny {
		t.Fatalf("type = %q, want canny", units[0].Type)
	}
	if units[0].ImagePath != "final.png" {
		t.Fatalf("image = %q, want final.png (last wins)", units[0].ImagePath)
	}
	if units[0].Weight != 0.6 {
		t.Fatalf("weight = %g, want 0.6", units[0].Weight)
	}
}

func TestBuildImageRequestControlSynonyms(t *testing.T) {
	req, err := BuildImageRequest("portrait", false, ImageRequestFields{
		ControlUnits: []driver.ControlNetUnit{
			{Type: driver.ImageControlPose, ImagePath: "pose.png", Preprocessor: "full", Weight: 0.9},
		},
	})
	if err != nil {
		t.Fatalf("BuildImageRequest: %v", err)
	}
	if req.Task != driver.ImageTaskControlNet {
		t.Fatalf("task = %q, want controlnet", req.Task)
	}
	if len(req.ControlUnits) != 1 || req.ControlUnits[0].Type != driver.ImageControlPose {
		t.Fatalf("units = %+v", req.ControlUnits)
	}
}
