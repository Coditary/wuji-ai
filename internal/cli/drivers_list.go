package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/coditary/wuji-core/pkg/netx"
)

// Keep in sync with built-in skips in wuji-core pkg/config/driver_endpoints.go.
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

func configuredPluginDriverIDs(app *App) []string {
	if app == nil || app.Config == nil {
		return nil
	}
	seen := make(map[string]bool)
	var ids []string
	add := func(id string) {
		if builtInDriverID(id) || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	for _, id := range app.Config.ConfiguredDriverIDs() {
		add(id)
	}
	for _, id := range reqpackInstalledDriverIDs() {
		add(id)
	}
	sort.Strings(ids)
	return ids
}

// pluginDriverStatus probes the host for a configured plugin (binary + gRPC port).
// Returns "" for built-ins or unknown IDs.
func pluginDriverStatus(cfg *config.Config, id string) string {
	if cfg == nil || builtInDriverID(id) {
		return ""
	}
	bin := cfg.DriverBinaryPath(id)
	if _, err := os.Stat(bin); err != nil {
		return "missing"
	}
	if netx.EndpointListening(cfg.DriverGRPCAddr(id)) {
		return "running"
	}
	return "stopped"
}

// inventoryPluginStatus is shown for plugins that are configured but not registered with Wuji.
func inventoryPluginStatus(cfg *config.Config, id string) string {
	switch pluginDriverStatus(cfg, id) {
	case "running":
		return "listening"
	case "stopped", "missing":
		return pluginDriverStatus(cfg, id)
	default:
		return ""
	}
}

func isConfiguredPluginDriver(app *App, id string) bool {
	for _, x := range configuredPluginDriverIDs(app) {
		if x == id {
			return true
		}
	}
	return false
}

func printInactivePluginDriverInfo(app *App, driverID string) error {
	status := inventoryPluginStatus(app.Config, driverID)
	if status == "" {
		return fmt.Errorf("driver %q not found", driverID)
	}
	fmt.Printf("ID:          %s\n", driverID)
	fmt.Printf("Status:      %s (not registered with Wuji)\n", status)
	fmt.Printf("Binary:      %s\n", app.Config.DriverBinaryPath(driverID))
	fmt.Printf("gRPC:        %s\n", app.Config.DriverGRPCAddr(driverID))
	fmt.Println()
	switch status {
	case "listening":
		fmt.Println("A plugin process is listening on gRPC but not attached to this Wuji runtime.")
		fmt.Println("Try: wuji list drivers --connect")
	case "missing":
		fmt.Println("Install the driver binary or fix drivers.<id>.driver_bin in config.")
	default:
		fmt.Println("The plugin will start on first use when auto_start_driver is enabled,")
		fmt.Println("or run: wuji list drivers --connect")
	}
	return nil
}

func printDriverInventory(app *App) error {
	if app == nil || app.Core == nil {
		return fmt.Errorf("core not available")
	}

	registered := make(map[string]driver.Info)
	for _, d := range app.Core.ListDrivers() {
		registered[d.ID] = d
	}

	var rows []string
	seen := make(map[string]bool)
	appendRow := func(id, status, name, version, remote, caps string) {
		if seen[id] {
			return
		}
		seen[id] = true
		rows = append(rows, strings.Join([]string{id, status, name, version, remote, caps}, "\t"))
	}

	activeIDs := make([]string, 0, len(registered))
	for id := range registered {
		activeIDs = append(activeIDs, id)
	}
	sort.Strings(activeIDs)
	for _, id := range activeIDs {
		d := registered[id]
		appendRow(d.ID, "active", d.Name, d.Version, remoteLabel(d.Remote), formatCapabilities(d.Capabilities))
	}

	for _, id := range configuredPluginDriverIDs(app) {
		if registered[id].ID != "" {
			continue
		}
		status := inventoryPluginStatus(app.Config, id)
		if status == "" {
			continue
		}
		appendRow(id, status, "-", "-", "yes", "-")
	}

	if len(rows) == 0 {
		fmt.Println("No drivers registered or configured.")
		return nil
	}

	sort.Strings(rows)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tNAME\tVERSION\tREMOTE\tCAPABILITIES")
	for _, row := range rows {
		fmt.Fprintln(w, row)
	}
	return w.Flush()
}

func remoteLabel(remote bool) string {
	if remote {
		return "yes"
	}
	return "no"
}

func formatCapabilities(caps []capability.Type) string {
	if len(caps) == 0 {
		return "-"
	}
	var parts []string
	for _, c := range caps {
		parts = append(parts, c.String())
	}
	return strings.Join(parts, ", ")
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
