package cli

import "testing"

func TestIsLocalCommand(t *testing.T) {
	tests := []struct {
		args []string
		want bool
	}{
		{[]string{"wuji"}, true},
		{[]string{"wuji", "--help"}, true},
		{[]string{"wuji", "-h"}, true},
		{[]string{"wuji", "help"}, true},
		{[]string{"wuji", "text", "--help"}, true},
		{[]string{"wuji", "catalog", "sync"}, true},
		{[]string{"wuji", "config", "show"}, true},
		{[]string{"wuji", "text", "hello"}, false},
		{[]string{"wuji", "list", "drivers"}, false},
		{[]string{"wuji", "list", "providers"}, true},
	}
	for _, tc := range tests {
		if got := isLocalCommand(tc.args); got != tc.want {
			t.Fatalf("isLocalCommand(%v) = %v, want %v", tc.args, got, tc.want)
		}
	}
}
