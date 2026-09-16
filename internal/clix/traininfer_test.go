package clix

import (
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestResolveImageMethodFromFlags(t *testing.T) {
	o := ImageTrainOpts{ClassToken: "sks person"}
	if ResolveImageMethod(o) != driver.ImageTrainMethodDreamBooth {
		t.Fatalf("got %s", ResolveImageMethod(o))
	}
	o = ImageTrainOpts{Method: driver.ImageTrainMethodLoRA, LoRARank: 16}
	if ResolveImageMethod(o) != driver.ImageTrainMethodLoRA {
		t.Fatal("explicit method should win")
	}
}

func TestResolveAudioMethodFromFlags(t *testing.T) {
	o := AudioTrainOpts{}
	if ResolveAudioMethod(o, true, false, "") != driver.AudioTrainMethodSFX {
		t.Fatal("expected sfx")
	}
	if ResolveAudioMethod(o, false, true, "") != driver.AudioTrainMethodSpeech {
		t.Fatal("expected speech")
	}
	if ResolveAudioMethod(o, false, false, "melody.mid") != driver.AudioTrainMethodMelody {
		t.Fatal("expected melody")
	}
}

func TestResolveTextMethodQLoRA(t *testing.T) {
	if ResolveTextMethod(TextTrainOpts{QLoRA: true}) != driver.TextTrainMethodQLoRA {
		t.Fatal("expected qlora")
	}
}

func TestResolveVideoMethodLoRA(t *testing.T) {
	if ResolveVideoMethod(VideoTrainOpts{LoRARank: 8}) != driver.VideoTrainMethodLoRA {
		t.Fatal("expected lora")
	}
}

func TestResolveMeshMethodLoRA(t *testing.T) {
	if ResolveMeshMethod(MeshTrainOpts{LoRARank: 4}) != driver.MeshTrainMethodLoRA {
		t.Fatal("expected lora")
	}
}

func TestResolveVoiceMethodDefault(t *testing.T) {
	if ResolveVoiceMethod(VoiceTrainOpts{}) != driver.VoiceTrainMethodRVC {
		t.Fatal("expected rvc")
	}
}

func TestFirstControlType(t *testing.T) {
	units := []driver.ControlNetUnit{{Type: ""}, {Type: driver.ImageControlDepth}}
	if FirstControlType(units) != driver.ImageControlDepth {
		t.Fatal("expected depth")
	}
}

func TestFormatMethodCatalog(t *testing.T) {
	out := FormatMethodCatalog([]driver.TrainMethodInfo{{
		Method: "lora", Description: "LoRA", Dataset: "pairs",
	}})
	if !strings.Contains(out, "lora") || !strings.Contains(out, "LoRA") {
		t.Fatalf("catalog=%q", out)
	}
}
