package clix

import "testing"

func TestJoinArgsEmpty(t *testing.T) {
	if JoinArgs(nil) != "" {
		t.Fatal("expected empty string")
	}
}

func TestJoinArgsSingle(t *testing.T) {
	if JoinArgs([]string{"only"}) != "only" {
		t.Fatal("single arg should pass through")
	}
}

func TestJoinArgsMultiple(t *testing.T) {
	got := JoinArgs([]string{"hello", "world"})
	if got != "hello world" {
		t.Fatalf("got %q", got)
	}
}
