package cli

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/core"
	"github.com/coditary/wuji-core/pkg/driver/dummy"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	cfg := &config.Config{DefaultDriver: dummy.DriverID}
	c, err := core.New(core.Config{AppConfig: cfg, DefaultDriverID: dummy.DriverID, Lazy: true})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}
	return &App{Core: c, Config: cfg, cleanup: func() { _ = c.Close() }}
}
