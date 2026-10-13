package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coditary/wuji-core/pkg/config"
)

func builtInDriverID(id string) bool {
	switch id {
	case "", "dummy", "ffmpeg", "local-rag", "local", "rag":
		return true
	default:
		return false
	}
}

// connectConfiguredDrivers starts or attaches plugin drivers so ListDrivers can see them.
func connectConfiguredDrivers(ctx context.Context, app *App) {
	if app == nil || app.Config == nil || app.Core == nil {
		return
	}
	for _, id := range app.Config.ConfiguredDriverIDs() {
		if builtInDriverID(id) {
			continue
		}
		if err := app.Core.EnsureDriver(ctx, id); err != nil {
			fmt.Fprintf(os.Stderr, "warning: driver %q: %v\n", id, err)
		}
	}
	// Also connect drivers installed via ReqPack when not yet referenced in capability_drivers.
	for _, id := range reqpackInstalledDriverIDs() {
		if builtInDriverID(id) {
			continue
		}
		if containsDriverID(app.Config.ConfiguredDriverIDs(), id) {
			continue
		}
		if config.ReqpackDriverBinaryPath(id) == "" {
			continue
		}
		if err := app.Core.EnsureDriver(ctx, id); err != nil {
			fmt.Fprintf(os.Stderr, "warning: driver %q: %v\n", id, err)
		}
	}
}

func containsDriverID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func reqpackInstalledDriverIDs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	dir := filepath.Join(home, ".local", "share", "wuji", "bin")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var ids []string
	const prefix = "wuji-driver-"
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) <= len(prefix) || name[:len(prefix)] != prefix {
			continue
		}
		ids = append(ids, name[len(prefix):])
	}
	return ids
}
