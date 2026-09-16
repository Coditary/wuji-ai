package cli

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("WUJI_EMBEDDED", "1")
	os.Exit(m.Run())
}
