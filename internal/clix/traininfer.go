package clix

import (
	"strings"

	"github.com/coditary/wuji-core/pkg/driver"
)

type TextTrainOpts struct {
	Method     driver.TextTrainMethod
	MethodFlag string
	QLoRA      bool
	LoRARank   int
}

type ImageTrainOpts struct {
	Method      driver.ImageTrainMethod
	MethodFlag  string
	ClassToken  string
	Token       string
	ControlType string
	LoRARank    int
}

type VideoTrainOpts struct {
	Method     driver.VideoTrainMethod
	MethodFlag string
	LoRARank   int
}

type AudioTrainOpts struct {
	Method     driver.AudioTrainMethod
	MethodFlag string
}

type MeshTrainOpts struct {
	Method     driver.MeshTrainMethod
	MethodFlag string
	LoRARank   int
}

type VoiceTrainOpts struct {
	Method     driver.VoiceTrainMethod
	MethodFlag string
}

func ResolveTextMethod(o TextTrainOpts) driver.TextTrainMethod {
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

func ResolveImageMethod(o ImageTrainOpts) driver.ImageTrainMethod {
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

func ResolveVideoMethod(o VideoTrainOpts) driver.VideoTrainMethod {
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

func ResolveAudioMethod(o AudioTrainOpts, sfx, speech bool, referencePath string) driver.AudioTrainMethod {
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

func ResolveMeshMethod(o MeshTrainOpts) driver.MeshTrainMethod {
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

func ResolveVoiceMethod(o VoiceTrainOpts) driver.VoiceTrainMethod {
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

func FirstControlType(units []driver.ControlNetUnit) driver.ImageControlType {
	for _, u := range units {
		if u.Type != "" {
			return u.Type
		}
	}
	return ""
}
