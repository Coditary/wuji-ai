package cli

import (
	"testing"
)

func TestParseImageScaleInput(t *testing.T) {
	v, requested, err := parseImageScaleInput("250%", false)
	if err != nil {
		t.Fatal(err)
	}
	if v != 2.5 || !requested {
		t.Fatalf("got scale=%g requested=%v", v, requested)
	}

	v, requested, err = parseImageScaleInput("", true)
	if err != nil {
		t.Fatal(err)
	}
	if v != 4 || !requested {
		t.Fatalf("default upscale: got scale=%g requested=%v", v, requested)
	}
}
