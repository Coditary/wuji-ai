package cli

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildMeshRequestGenerate(t *testing.T) {
	req, err := buildMeshRequest("low poly crate", false, meshRequestFields{format: "glb"})
	if err != nil {
		t.Fatalf("buildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskGenerate {
		t.Fatalf("task = %q, want generate", req.Task)
	}
}

func TestBuildMeshRequestImageToMesh(t *testing.T) {
	req, err := buildMeshRequest("", false, meshRequestFields{imagePath: "ref.png"})
	if err != nil {
		t.Fatalf("buildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskImageToMesh {
		t.Fatalf("task = %q, want i2m", req.Task)
	}
}

func TestBuildMeshRequestExplicitTask(t *testing.T) {
	req, err := buildMeshRequest("x", false, meshRequestFields{
		taskExplicit: "repair",
		meshPath:     "scan.glb",
	})
	if err != nil {
		t.Fatalf("buildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskRepair {
		t.Fatalf("task = %q, want repair", req.Task)
	}
}

func TestBuildMeshRequestScene(t *testing.T) {
	req, err := buildMeshRequest("medieval room", false, meshRequestFields{sceneRequested: true})
	if err != nil {
		t.Fatalf("buildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskScene {
		t.Fatalf("task = %q, want scene", req.Task)
	}
}

func TestBuildMeshRequestMode(t *testing.T) {
	req, err := buildMeshRequest("knight", false, meshRequestFields{mode: "human"})
	if err != nil {
		t.Fatalf("buildMeshRequest: %v", err)
	}
	if req.Mode != driver.MeshModeHuman {
		t.Fatalf("mode = %q, want human", req.Mode)
	}
}

func TestBuildMeshRequestUpscale(t *testing.T) {
	req, err := buildMeshRequest("", false, meshRequestFields{
		meshPath:         "hero.glb",
		upscaleRequested: true,
		scale:            2,
		scaleSet:         true,
	})
	if err != nil {
		t.Fatalf("buildMeshRequest: %v", err)
	}
	if req.Task != driver.MeshTaskUpscale {
		t.Fatalf("task = %q, want upscale", req.Task)
	}
}
