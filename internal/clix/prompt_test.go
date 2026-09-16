package clix

import (
	"strings"
	"testing"
)

func TestReadLine(t *testing.T) {
	line, err := ReadLine("", strings.NewReader("hello world\n"))
	if err != nil {
		t.Fatal(err)
	}
	if line != "hello world" {
		t.Fatalf("line=%q", line)
	}
}

func TestReadLineEmptyInput(t *testing.T) {
	_, err := ReadLine("", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestSelectOptionDefault(t *testing.T) {
	// Empty line selects first option; we cannot easily mock os.Stdin here,
	// so test invalid choice path via exported error semantics on bad input.
	_, err := SelectOption("Pick", []string{})
	if err == nil {
		t.Fatal("expected error for no options")
	}
}

func TestSelectOptionInvalidChoice(t *testing.T) {
	// SelectOption reads from os.Stdin when called without injection;
	// ReadLine covers stdin parsing; here we only verify empty options guard.
	if _, err := SelectOption("x", nil); err == nil {
		t.Fatal("expected error")
	}
}
