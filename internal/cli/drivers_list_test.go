package cli

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/core"
	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/coditary/wuji-core/pkg/scheduler"
)

func TestBuiltInDriverID(t *testing.T) {
	for _, id := range []string{"dummy", "ffmpeg", "local-rag", "local", "rag", ""} {
		if !builtInDriverID(id) {
			t.Fatalf("expected built-in %q", id)
		}
	}
	if builtInDriverID("llama") {
		t.Fatal("llama should not be built-in")
	}
}

func TestContainsDriverID(t *testing.T) {
	ids := []string{"a", "b"}
	if !containsDriverID(ids, "b") {
		t.Fatal("expected b")
	}
	if containsDriverID(ids, "c") {
		t.Fatal("unexpected c")
	}
}

func TestReqpackInstalledDriverIDs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	bin := home + "/.local/share/wuji/bin"
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin+"/wuji-driver-llama", nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin+"/not-a-driver", nil, 0o644); err != nil {
		t.Fatal(err)
	}

	got := reqpackInstalledDriverIDs()
	if len(got) != 1 || got[0] != "llama" {
		t.Fatalf("got %#v", got)
	}
}

func TestPluginDriverStatus(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "wuji-driver-llama")
	if err := os.WriteFile(bin, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Root: dir,
		CapabilityDrivers: config.CapabilityDrivers{"text": "llama"},
		Drivers: config.DriversConfig{
			Registry: map[string]config.DriverSection{
				"llama": {DriverBin: bin},
			},
		},
	}
	if got := pluginDriverStatus(cfg, "llama"); got != "stopped" {
		t.Fatalf("stopped: got %q", got)
	}
	if got := pluginDriverStatus(cfg, "dummy"); got != "" {
		t.Fatalf("built-in: got %q", got)
	}
	if got := pluginDriverStatus(cfg, "missing"); got != "missing" {
		t.Fatalf("missing binary: got %q", got)
	}
}

func TestInventoryPluginStatus_Listening(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "wuji-driver-llama")
	if err := os.WriteFile(bin, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := &config.Config{
		Root: dir,
		CapabilityDrivers: config.CapabilityDrivers{"text": "custom"},
		Drivers: config.DriversConfig{
			Registry: map[string]config.DriverSection{
				"custom": {DriverBin: bin, GRPCAddr: ln.Addr().String()},
			},
		},
	}
	if got := inventoryPluginStatus(cfg, "custom"); got != "listening" {
		t.Fatalf("got %q, want listening", got)
	}
}

func TestPrintDriverInventory_ConfiguredPlugin(t *testing.T) {
	app := newTestApp(t)
	dir := t.TempDir()
	bin := filepath.Join(dir, "wuji-driver-llama")
	if err := os.WriteFile(bin, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	app.Config.CapabilityDrivers = config.CapabilityDrivers{"text": "llama"}
	app.Config.Drivers.Registry = map[string]config.DriverSection{
		"llama": {DriverBin: bin},
	}

	out, err := captureStdout(func() error {
		return printDriverInventory(app)
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"STATUS", "dummy", "active", "llama", "stopped"} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected %q in output:\n%s", part, out)
		}
	}
}

func TestConnectConfiguredDrivers_EnsurePluginOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	rec := &ensureDriverRecorder{}
	auto := true
	cfg := &config.Config{
		CapabilityDrivers: config.CapabilityDrivers{"text": "llama", "image": "dummy"},
		Drivers: config.DriversConfig{
			Registry: map[string]config.DriverSection{
				"llama": {DriverBin: "/tmp/wuji-driver-llama", AutoStartDriver: &auto},
			},
		},
	}
	app := &App{Core: rec, Config: cfg}
	connectConfiguredDrivers(context.Background(), app)

	if len(rec.calls) != 1 || rec.calls[0] != "llama" {
		t.Fatalf("EnsureDriver calls: %#v", rec.calls)
	}
}

// ensureDriverRecorder records EnsureDriver for connectConfiguredDrivers tests.
type ensureDriverRecorder struct {
	calls []string
}

func (e *ensureDriverRecorder) EnsureDriver(ctx context.Context, driverID string) error {
	e.calls = append(e.calls, driverID)
	return nil
}

func (e *ensureDriverRecorder) ListDrivers() []driver.Info { return nil }
func (e *ensureDriverRecorder) Close() error               { return nil }
func (e *ensureDriverRecorder) DefaultDriverID() string    { return "dummy" }

func (e *ensureDriverRecorder) GenerateText(context.Context, string, driver.TextRequest) (*driver.TextResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) GenerateTextStream(context.Context, string, driver.TextRequest, func(string) error) (*driver.TextResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) GenerateImage(context.Context, string, driver.ImageRequest) (*driver.ImageResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) GenerateVideo(context.Context, string, driver.VideoRequest) (*driver.VideoResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) GenerateAudio(context.Context, string, driver.AudioRequest) (*driver.AudioResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) GenerateMesh(context.Context, string, driver.MeshRequest) (*driver.MeshResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) ManageDataset(context.Context, string, driver.DatasetRequest) (*driver.DatasetResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) TrainText(context.Context, string, driver.TextTrainRequest) (*driver.TrainResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) TrainImage(context.Context, string, driver.ImageTrainRequest) (*driver.TrainResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) TrainVideo(context.Context, string, driver.VideoTrainRequest) (*driver.TrainResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) TrainAudio(context.Context, string, driver.AudioTrainRequest) (*driver.TrainResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) TrainMesh(context.Context, string, driver.MeshTrainRequest) (*driver.TrainResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) TrainVoice(context.Context, string, driver.VoiceTrainRequest) (*driver.TrainResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) RunRAG(context.Context, string, driver.RAGRequest) (data.DataShape, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) ProduceData(context.Context, string, driver.DataRequest) (data.DataShape, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) ProduceDecision(context.Context, string, driver.DecisionRequest) (*driver.DecisionResponse, error) {
	return nil, nil
}
func (e *ensureDriverRecorder) ConnectRemote(context.Context, string, bool) error { return nil }
func (e *ensureDriverRecorder) ResourceStatus() (scheduler.PersistedState, error) {
	return scheduler.PersistedState{}, nil
}
func (e *ensureDriverRecorder) ResourcesConfig() config.EffectiveResources {
	return config.EffectiveResources{}
}
func (e *ensureDriverRecorder) LoadInference(context.Context, string, string, capability.Type) error {
	return nil
}
func (e *ensureDriverRecorder) UnloadInferenceDriver(context.Context, string) error { return nil }
func (e *ensureDriverRecorder) UnloadAllInference(context.Context) error            { return nil }

var _ core.Runtime = (*ensureDriverRecorder)(nil)
