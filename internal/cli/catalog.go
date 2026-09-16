package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	wujicfg "github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/data"
	"github.com/coditary/wuji-core/pkg/driver"
	"github.com/coditary/wuji-core/pkg/mcp"
	"github.com/coditary/wuji-core/pkg/modelformat"
)

func ragStoreRoot(app *App) string {
	if app.Config.Root != "" {
		return app.Config.Root
	}
	return "."
}

func printListOverview(app *App) error {
	drivers := app.Core.ListDrivers()
	mcpMgr := mcp.NewManager(app.Config.Root)
	mcpList, _ := mcpMgr.ListStatus(context.Background())

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TOPIC\tCOMMAND")
	fmt.Fprintf(w, "drivers (%d)\t%s\n", len(drivers), "wuji list drivers")
	fmt.Fprintf(w, "mcp servers (%d)\t%s\n", len(mcpList), "wuji list mcp")
	fmt.Fprintf(w, "datasets\t%s\n", "wuji list datasets")
	fmt.Fprintf(w, "loras\t%s\n", "wuji list loras")
	fmt.Fprintf(w, "providers (catalog)\t%s\n", "wuji list providers")
	fmt.Fprintf(w, "models (catalog)\t%s\n", "wuji list models")
	fmt.Fprintf(w, "labs (catalog)\t%s\n", "wuji list labs")
	fmt.Fprintf(w, "routes (config)\t%s\n", "wuji list routes")
	fmt.Fprintf(w, "capabilities\t%s\n", "wuji list capabilities")
	fmt.Fprintf(w, "formats\t%s\n", "wuji list formats")
	fmt.Fprintf(w, "rag collections\t%s\n", "wuji list collections")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "DETAILS\tCOMMAND")
	fmt.Fprintln(w, "driver details\twuji info driver <id>")
	fmt.Fprintln(w, "mcp server details\twuji info mcp <name>")
	fmt.Fprintln(w, "effective config\twuji info config")
	fmt.Fprintln(w, "rag collection\twuji info collection <name>")
	return w.Flush()
}

func printInfoOverview() error {
	fmt.Println("Show detailed information about Wuji components.")
	fmt.Println()
	fmt.Println("  wuji info driver <id>       driver capabilities, tasks, and formats")
	fmt.Println("  wuji info mcp <name>        MCP server configuration and status")
	fmt.Println("  wuji info config            effective .wuji/config.yaml")
	fmt.Println("  wuji info collection <name> RAG collection metadata")
	return nil
}

func printDriverList(app *App) error {
	drivers := app.Core.ListDrivers()
	if len(drivers) == 0 {
		fmt.Println("No drivers registered.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tVERSION\tREMOTE\tCAPABILITIES")
	for _, d := range drivers {
		remote := "no"
		if d.Remote {
			remote = "yes"
		}

		caps := ""
		for i, c := range d.Capabilities {
			if i > 0 {
				caps += ", "
			}
			caps += c.String()
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.ID, d.Name, d.Version, remote, caps)
	}
	return w.Flush()
}

func printDriverInfo(app *App, driverID string) error {
	for _, d := range app.Core.ListDrivers() {
		if d.ID != driverID {
			continue
		}

		fmt.Printf("ID:          %s\n", d.ID)
		fmt.Printf("Name:        %s\n", d.Name)
		fmt.Printf("Version:     %s\n", d.Version)
		fmt.Printf("Description: %s\n", d.Description)
		if d.Remote {
			fmt.Printf("Endpoint:    %s\n", d.Endpoint)
		}

		fmt.Println("\nCapabilities:")
		for _, c := range d.Capabilities {
			fmt.Printf("  - %s\n", c)
		}

		if len(d.ImageTasks) > 0 {
			fmt.Println("\nImage tasks:")
			for _, task := range d.ImageTasks {
				fmt.Printf("  - %s\n", task)
			}
		}

		if len(d.VideoTasks) > 0 {
			fmt.Println("\nVideo tasks:")
			for _, task := range d.VideoTasks {
				fmt.Printf("  - %s\n", task)
			}
		}

		if len(d.DatasetTasks) > 0 {
			fmt.Println("\nDataset tasks:")
			for _, task := range d.DatasetTasks {
				fmt.Printf("  - %s\n", task)
			}
		}

		if len(d.AudioTasks) > 0 {
			fmt.Println("\nAudio tasks:")
			for _, task := range d.AudioTasks {
				fmt.Printf("  - %s\n", task)
			}
		}

		if len(d.MeshTasks) > 0 {
			fmt.Println("\nMesh tasks:")
			for _, task := range d.MeshTasks {
				fmt.Printf("  - %s\n", task)
			}
		}

		fmt.Println("\nSupported source formats:")
		if len(d.FormatSupport) == 0 {
			fmt.Println("  (none declared)")
			return nil
		}
		for _, c := range capability.All() {
			formats := d.FormatSupport.FormatsFor(c)
			if len(formats) == 0 {
				continue
			}
			names := make([]string, 0, len(formats))
			for _, f := range formats {
				names = append(names, f.String())
			}
			fmt.Printf("  %s: %s\n", c, strings.Join(names, ", "))
		}
		return nil
	}
	return fmt.Errorf("driver %q not found", driverID)
}

func printCapabilityList() {
	fmt.Println("Supported capabilities:")
	for _, c := range capability.All() {
		fmt.Printf("  - %s\n", c)
	}
}

func printFormatList() {
	fmt.Println("Known source formats:")
	for _, f := range modelformat.All() {
		fmt.Printf("  %-18s %s\n", f, f.Description())
	}
}

func printEffectiveConfig(cfg *wujicfg.Config) error {
	vllm := cfg.ResolvedVLLM()
	llama := cfg.ResolvedLlama()
	a1111 := cfg.ResolvedA1111()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ROOT\t%s\n", cfg.Root)
	if cfg.DefaultDriver != "" {
		fmt.Fprintf(w, "DEFAULT_DRIVER\t%s\n", cfg.DefaultDriver)
	}
	if cfg.DefaultProvider != "" {
		fmt.Fprintf(w, "DEFAULT_PROVIDER\t%s\n", cfg.DefaultProvider)
	}
	if len(cfg.CapabilityDrivers) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "CAPABILITY_DRIVERS")
		for cap, driverID := range cfg.CapabilityDrivers {
			fmt.Fprintf(w, "  %s\t%s\n", cap, driverID)
		}
	}
	res := cfg.ResolvedResources()
	if res.Enabled() {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "RESOURCES")
		if res.TotalVRAMMB > 0 {
			fmt.Fprintf(w, "  total_vram_mb\t%d\n", res.TotalVRAMMB)
		}
		if res.TotalRAMMB > 0 {
			fmt.Fprintf(w, "  total_ram_mb\t%d\n", res.TotalRAMMB)
		}
		fmt.Fprintf(w, "  evict_and_restore\t%t\n", res.EvictAndRestore)
		fmt.Fprintf(w, "  lazy_restore\t%t\n", res.LazyRestore)
		fmt.Fprintf(w, "  unload_transient_models\t%t\n", res.UnloadTransientModels)
		if res.IdleShutdownSeconds > 0 {
			fmt.Fprintf(w, "  idle_shutdown_seconds\t%d\n", res.IdleShutdownSeconds)
		} else {
			fmt.Fprintln(w, "  idle_shutdown_seconds\t0 (disabled)")
		}
	}
	if loraText := describeLoRAConfig(cfg.ResolvedLoRAs()); loraText != "" {
		fmt.Fprintln(w, "")
		fmt.Fprint(w, loraText)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "VLLM")
	fmt.Fprintf(w, "  vllm_bin\t%s\n", vllm.VllmBin)
	fmt.Fprintf(w, "  default_model\t%s\n", displayOrDash(vllm.DefaultModel))
	fmt.Fprintf(w, "  host\t%s\n", vllm.Host)
	fmt.Fprintf(w, "  port\t%d\n", vllm.Port)
	fmt.Fprintf(w, "  api_base\t%s\n", displayOrDash(vllm.APIBase))
	fmt.Fprintf(w, "  grpc_addr\t%s\n", vllm.GRPCAddr)
	fmt.Fprintf(w, "  timeout_seconds\t%d\n", vllm.TimeoutSeconds)
	fmt.Fprintf(w, "  startup_timeout_seconds\t%d\n", vllm.StartupTimeoutSeconds)
	fmt.Fprintf(w, "  manage_server\t%t\n", vllm.ManageServerEnabled())
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "LLAMA")
	fmt.Fprintf(w, "  server_bin\t%s\n", llama.ServerBin)
	fmt.Fprintf(w, "  models_dir\t%s\n", llama.ModelsDir)
	fmt.Fprintf(w, "  default_model\t%s\n", displayOrDash(llama.DefaultModel))
	fmt.Fprintf(w, "  inference_host\t%s\n", llama.InferenceHost)
	fmt.Fprintf(w, "  inference_port\t%d\n", llama.InferencePort)
	fmt.Fprintf(w, "  grpc_addr\t%s\n", llama.GRPCAddr)
	fmt.Fprintf(w, "  ollama_api\t%s\n", llama.OllamaAPI)
	fmt.Fprintf(w, "  ollama_think\t%t\n", llama.OllamaThinkEnabled())
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "A1111")
	fmt.Fprintf(w, "  host\t%s\n", a1111.Host)
	fmt.Fprintf(w, "  port\t%d\n", a1111.Port)
	fmt.Fprintf(w, "  api_base\t%s\n", displayOrDash(a1111.APIBase))
	fmt.Fprintf(w, "  default_model\t%s\n", displayOrDash(a1111.DefaultModel))
	fmt.Fprintf(w, "  default_video_model\t%s\n", displayOrDash(a1111.DefaultVideoModel))
	fmt.Fprintf(w, "  grpc_addr\t%s\n", a1111.GRPCAddr)
	fmt.Fprintf(w, "  timeout_seconds\t%d\n", a1111.TimeoutSeconds)
	fmt.Fprintf(w, "  output_dir\t%s\n", a1111.OutputDir)
	fmt.Fprintf(w, "  video_output_dir\t%s\n", a1111.VideoOutputDir)
	fmt.Fprintf(w, "  video_width\t%d\n", a1111.VideoWidth)
	fmt.Fprintf(w, "  video_height\t%d\n", a1111.VideoHeight)
	fmt.Fprintf(w, "  video_steps\t%d\n", a1111.VideoSteps)
	fmt.Fprintf(w, "  video_cfg_scale\t%d\n", a1111.VideoCFGScale)
	fmt.Fprintf(w, "  video_sampler\t%s\n", displayOrDash(a1111.VideoSampler))
	fmt.Fprintf(w, "  video_fps\t%d\n", a1111.VideoFPS)
	return w.Flush()
}

func printLoRAList(cfg *wujicfg.Config) error {
	loraCfg := cfg.ResolvedLoRAs()
	if len(loraCfg.Library) == 0 && len(loraCfg.Default) == 0 {
		fmt.Println("No LoRAs configured.")
		fmt.Println("Hint: wuji config set lora <alias> <path>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if loraCfg.DefaultWeight > 0 && loraCfg.DefaultWeight != 1 {
		fmt.Fprintf(w, "DEFAULT_WEIGHT\t%g\n", loraCfg.DefaultWeight)
	}
	if len(loraCfg.Default) > 0 {
		fmt.Fprintf(w, "DEFAULT\t%s\n", strings.Join(loraCfg.Default, ", "))
	}

	names := make([]string, 0, len(loraCfg.Library))
	for name := range loraCfg.Library {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprintln(w, "ALIAS\tPATH\tWEIGHT")
	for _, name := range names {
		entry := loraCfg.Library[name]
		weight := "-"
		if entry.Weight > 0 {
			weight = fmt.Sprintf("%g", entry.Weight)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, entry.Path, weight)
	}
	return w.Flush()
}

func providerModelForAPI(entry wujicfg.APIEntry) string {
	if entry.Text != nil && entry.Text.Model != "" {
		return entry.Text.Model
	}
	if entry.Image != nil && entry.Image.Model != "" {
		return entry.Image.Model
	}
	if entry.Audio != nil && entry.Audio.Model != "" {
		return entry.Audio.Model
	}
	return ""
}

func printRouteList(cfg *wujicfg.Config) error {
	models := map[string]string{}
	for id, entry := range cfg.APIs {
		models[id] = providerModelForAPI(entry)
	}
	for id, entry := range cfg.Providers {
		if models[id] == "" {
			models[id] = entry.Model
		}
	}
	if len(models) == 0 {
		fmt.Println("No routes configured.")
		fmt.Println("Hint: add entries under apis: or providers: in .wuji/config.yaml")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defaultID := cfg.DefaultProvider
	if defaultID == "" {
		defaultID = cfg.DefaultDriver
	}
	if defaultID != "" {
		fmt.Fprintf(w, "DEFAULT\t%s\n", defaultID)
		fmt.Fprintln(w, "")
	}
	fmt.Fprintln(w, "ID\tMODEL")
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		model := models[name]
		if model == "" {
			model = "-"
		}
		fmt.Fprintf(w, "%s\t%s\n", name, model)
	}
	return w.Flush()
}

func printDatasetList(app *App, cmd *cobra.Command) error {
	resp, err := app.Core.ManageDataset(context.Background(), app.resolveDriver(cmd, capability.DatasetMgmt), driver.DatasetRequest{
		Task: driver.DatasetTaskList,
	})
	if err != nil {
		return err
	}
	if len(resp.Datasets) == 0 {
		fmt.Println("No datasets found.")
		return nil
	}
	for _, ds := range resp.Datasets {
		fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%d bytes\t%s (%d versions)\n",
			ds.ID, ds.Name, ds.Path, ds.Size, ds.LatestVersion, ds.VersionCount)
	}
	return nil
}

func printRAGCollections(app *App, cmd *cobra.Command) error {
	shape, err := app.Core.RunRAG(context.Background(), app.resolveDriver(cmd, capability.RAG), driver.RAGRequest{
		Task:      driver.RAGTaskList,
		StoreRoot: ragStoreRoot(app),
	})
	if err != nil {
		return err
	}
	rs, ok := shape.(*data.RecordSet)
	if !ok || rs == nil {
		return fmt.Errorf("unexpected rag list response")
	}
	if len(rs.Records) == 0 {
		fmt.Println("No RAG collections found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "COLLECTION\tCHUNKS\tDIMS\tEMBED_MODEL\tUPDATED")
	for _, rec := range rs.Records {
		if rec.Role != "collection" {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			fieldString(rec, "collection"),
			fieldString(rec, "chunk_count"),
			fieldString(rec, "dims"),
			fieldString(rec, "embed_model"),
			fieldString(rec, "updated_at"),
		)
	}
	return w.Flush()
}

func printRAGCollectionInfo(app *App, cmd *cobra.Command, name string) error {
	shape, err := app.Core.RunRAG(context.Background(), app.resolveDriver(cmd, capability.RAG), driver.RAGRequest{
		Task:       driver.RAGTaskInfo,
		StoreRoot:  ragStoreRoot(app),
		Collection: name,
	})
	if err != nil {
		return err
	}
	rs, ok := shape.(*data.RecordSet)
	if !ok || rs == nil {
		return fmt.Errorf("unexpected rag info response")
	}
	for _, rec := range rs.Records {
		if rec.Role != "collection" {
			continue
		}
		fmt.Printf("Collection:  %s\n", fieldString(rec, "collection"))
		fmt.Printf("Embed model: %s\n", fieldString(rec, "embed_model"))
		fmt.Printf("Dimensions:  %s\n", fieldString(rec, "dims"))
		fmt.Printf("Chunks:      %s\n", fieldString(rec, "chunk_count"))
		if v := fieldString(rec, "chunk_size"); v != "" {
			fmt.Printf("Chunk size:  %s\n", v)
		}
		if v := fieldString(rec, "overlap"); v != "" {
			fmt.Printf("Overlap:     %s\n", v)
		}
		if v := fieldString(rec, "updated_at"); v != "" {
			fmt.Printf("Updated:     %s\n", v)
		}
		return nil
	}
	return fmt.Errorf("collection %q not found", name)
}
