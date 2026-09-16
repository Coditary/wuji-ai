package cli

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestResolveImageMethodFromFlags(t *testing.T) {
	o := imageTrainOpts{ClassToken: "sks person"}
	if resolveImageMethod(o) != driver.ImageTrainMethodDreamBooth {
		t.Fatalf("got %s", resolveImageMethod(o))
	}
	o = imageTrainOpts{Method: driver.ImageTrainMethodLoRA, LoRARank: 16}
	if resolveImageMethod(o) != driver.ImageTrainMethodLoRA {
		t.Fatalf("explicit method should win")
	}
}

func TestResolveAudioMethodFromFlags(t *testing.T) {
	o := audioTrainOpts{}
	if resolveAudioMethod(o, true, false, "") != driver.AudioTrainMethodSFX {
		t.Fatal("expected sfx")
	}
	if resolveAudioMethod(o, false, true, "") != driver.AudioTrainMethodSpeech {
		t.Fatal("expected speech")
	}
}
