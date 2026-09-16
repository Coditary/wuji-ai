package cli

import (
	"strings"

	"github.com/coditary/wuji-core/pkg/driver"
)

func resolveTextMethod(o textTrainOpts) driver.TextTrainMethod {
	if o.Method != "" {
		return o.Method
	}
	if o.MethodFlag != "" {
		m, err := driver.ParseTextTrainMethod(o.MethodFlag)
		if err == nil {
			return m
		}
	}
	if o.QLoRA {
		return driver.TextTrainMethodQLoRA
	}
	if o.LoRARank > 0 {
		return driver.TextTrainMethodLoRA
	}
	return driver.TextTrainMethodFull
}

func resolveImageMethod(o imageTrainOpts) driver.ImageTrainMethod {
	if o.Method != "" {
		return o.Method
	}
	if o.MethodFlag != "" {
		m, err := driver.ParseImageTrainMethod(o.MethodFlag)
		if err == nil {
			return m
		}
	}
	if o.ClassToken != "" {
		return driver.ImageTrainMethodDreamBooth
	}
	if o.Token != "" {
		return driver.ImageTrainMethodTextualInversion
	}
	if o.ControlType != "" {
		return driver.ImageTrainMethodControlNet
	}
	if o.LoRARank > 0 {
		return driver.ImageTrainMethodLoRA
	}
	return driver.ImageTrainMethodFull
}

func resolveVideoMethod(o videoTrainOpts) driver.VideoTrainMethod {
	if o.Method != "" {
		return o.Method
	}
	if o.MethodFlag != "" {
		m, err := driver.ParseVideoTrainMethod(o.MethodFlag)
		if err == nil {
			return m
		}
	}
	if o.LoRARank > 0 {
		return driver.VideoTrainMethodLoRA
	}
	return driver.VideoTrainMethodFull
}

func resolveAudioMethod(o audioTrainOpts, sfx, speech bool, referencePath string) driver.AudioTrainMethod {
	if o.Method != "" {
		return o.Method
	}
	if o.MethodFlag != "" {
		m, err := driver.ParseAudioTrainMethod(o.MethodFlag)
		if err == nil {
			return m
		}
	}
	if speech {
		return driver.AudioTrainMethodSpeech
	}
	if sfx {
		return driver.AudioTrainMethodSFX
	}
	if strings.TrimSpace(referencePath) != "" {
		return driver.AudioTrainMethodMelody
	}
	return driver.AudioTrainMethodMusic
}

func resolveMeshMethod(o meshTrainOpts) driver.MeshTrainMethod {
	if o.Method != "" {
		return o.Method
	}
	if o.MethodFlag != "" {
		m, err := driver.ParseMeshTrainMethod(o.MethodFlag)
		if err == nil {
			return m
		}
	}
	if o.LoRARank > 0 {
		return driver.MeshTrainMethodLoRA
	}
	return driver.MeshTrainMethodFull
}

func resolveVoiceMethod(o voiceTrainOpts) driver.VoiceTrainMethod {
	if o.Method != "" {
		return o.Method
	}
	if o.MethodFlag != "" {
		m, err := driver.ParseVoiceTrainMethod(o.MethodFlag)
		if err == nil {
			return m
		}
	}
	return driver.VoiceTrainMethodRVC
}

func firstControlType(units []driver.ControlNetUnit) driver.ImageControlType {
	for _, u := range units {
		if u.Type != "" {
			return u.Type
		}
	}
	return ""
}
