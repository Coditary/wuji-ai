package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	wujicfg "github.com/coditary/wuji-core/pkg/config"
	"github.com/coditary/wuji-core/pkg/mcp"
)

func newMcpCmd(app *App) *cobra.Command {
	root := app.Config.Root
	mgr := mcp.NewManager(root)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Model Context Protocol) servers",
		Long: `Start, stop, and manage MCP servers defined in .wuji/mcp.json.

The config file uses the same format as Cursor's mcp.json (mcpServers).

Transport modes for "wuji mcp start":
  --stdio   Foreground MCP proxy (for OpenCode: type local)
  --http    Background Streamable HTTP gateway (for OpenCode: type remote)
  --sse     Background SSE gateway
  (default) Background daemon with logs only

OpenCode integration:
  wuji mcp sync opencode              # write opencode.json (--stdio entries)
  wuji mcp sync opencode --http       # remote URLs after starting gateways

Import from Cursor:
  wuji mcp --import                        # .cursor/mcp.json or ~/.cursor/mcp.json
  wuji mcp --import ~/.cursor/mcp.json
  wuji mcp --import .cursor/mcp.json list  # import, then list`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("import") {
				return nil
			}
			importPath, _ := cmd.Flags().GetString("import")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			return runMCPImport(root, importPath, overwrite)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return printMCPList(cmd.Context(), mgr)
		},
	}
	cmd.PersistentFlags().String("import", "", "merge servers from mcp.json (default: auto-detect Cursor config)")
	cmd.PersistentFlags().Bool("overwrite", false, "replace existing server definitions when using --import")

	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add or update an MCP server definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command, _ := cmd.Flags().GetString("command")
			url, _ := cmd.Flags().GetString("url")
			cwd, _ := cmd.Flags().GetString("cwd")
			disabled, _ := cmd.Flags().GetBool("disabled")
			argsList, _ := cmd.Flags().GetStringArray("arg")
			envList, _ := cmd.Flags().GetStringArray("env")

			def := wujicfg.MCPServerDef{
				Command:  command,
				Args:     argsList,
				Env:      parseEnvPairs(envList),
				Cwd:      cwd,
				URL:      url,
				Disabled: disabled,
			}
			if err := wujicfg.AddMCPServer(root, args[0], def); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Added MCP server %q to .wuji/mcp.json\n", args[0])
			return nil
		},
	}
	addCmd.Flags().String("command", "", "executable for stdio MCP server (e.g. npx, uvx)")
	addCmd.Flags().StringArray("arg", nil, "command argument (repeatable)")
	addCmd.Flags().StringArray("env", nil, "environment variable KEY=VALUE (repeatable)")
	addCmd.Flags().String("url", "", "remote MCP server URL (HTTP/SSE)")
	addCmd.Flags().String("cwd", "", "working directory for the subprocess")
	addCmd.Flags().Bool("disabled", false, "mark server as disabled")

	cmd.AddCommand(
		addCmd,
		&cobra.Command{
			Use:   "remove <name>",
			Short: "Remove an MCP server definition",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := mcp.NewManager(root).Stop(args[0]); err != nil {
					_ = err
				}
				if err := wujicfg.RemoveMCPServer(root, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Removed MCP server %q\n", args[0])
				return nil
			},
		},
		&cobra.Command{
			Use:        "list",
			Short:      "List configured MCP servers and their status",
			Aliases:    []string{"ls"},
			Deprecated: "use `wuji list mcp` instead",
			RunE: func(cmd *cobra.Command, args []string) error {
				return printMCPList(cmd.Context(), mgr)
			},
		},
		&cobra.Command{
			Use:        "status [name]",
			Short:      "Show detailed status for one or all MCP servers",
			Deprecated: "use `wuji info mcp` instead",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return printMCPList(cmd.Context(), mgr)
				}
				st, err := mgr.Status(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				return printMCPDetail(st)
			},
		},
	)

	startCmd := &cobra.Command{
		Use:   "start [name]",
		Short: "Start an MCP server (or all with --all)",
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			stdio, _ := cmd.Flags().GetBool("stdio")
			http, _ := cmd.Flags().GetBool("http")
			sse, _ := cmd.Flags().GetBool("sse")
			ctx := cmd.Context()

			if stdio && (http || sse) {
				return fmt.Errorf("use only one of --stdio, --http, or --sse")
			}
			if http && sse {
				return fmt.Errorf("use only one of --http or --sse")
			}
			if stdio && (all || len(args) == 0) {
				return fmt.Errorf("--stdio requires a server name")
			}
			if stdio {
				return mgr.RunStdio(ctx, args[0])
			}

			gwOpts := parseGatewayFlags(cmd)
			if http || sse {
				transport := mcp.TransportHTTP
				if sse {
					transport = mcp.TransportSSE
				}
				if all || len(args) == 0 {
					started, gateways, errs := mgr.StartAllGateways(ctx, transport, gwOpts)
					for i, name := range started {
						fmt.Fprintf(os.Stdout, "Started %s at %s\n", name, gateways[i].URL)
					}
					for _, err := range errs {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					}
					if len(errs) > 0 {
						return fmt.Errorf("%d server(s) failed to start", len(errs))
					}
					if len(started) == 0 {
						fmt.Fprintln(os.Stdout, "No MCP servers configured.")
					}
					return nil
				}
				gw, err := mgr.StartGateway(ctx, args[0], transport, gwOpts, 0)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Started %s at %s\n", args[0], gw.URL)
				return nil
			}

			if all || len(args) == 0 {
				started, errs := mgr.StartAll(ctx)
				for _, name := range started {
					fmt.Fprintf(os.Stdout, "Started %s\n", name)
				}
				for _, err := range errs {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				if len(errs) > 0 {
					return fmt.Errorf("%d server(s) failed to start", len(errs))
				}
				if len(started) == 0 {
					fmt.Fprintln(os.Stdout, "No MCP servers configured.")
				}
				return nil
			}

			if err := mgr.Start(ctx, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Started %s\n", args[0])
			return nil
		},
	}
	startCmd.Flags().Bool("all", false, "start all enabled MCP servers")
	startCmd.Flags().Bool("stdio", false, "run as foreground stdio MCP proxy (for OpenCode type: local)")
	startCmd.Flags().Bool("http", false, "expose via Streamable HTTP gateway in background")
	startCmd.Flags().Bool("sse", false, "expose via SSE gateway in background")
	startCmd.Flags().String("host", "127.0.0.1", "gateway bind host (--http/--sse)")
	startCmd.Flags().Int("port", 0, "gateway port (--http/--sse); 0 = auto")
	startCmd.Flags().Int("port-base", 9333, "first auto port when starting multiple gateways")
	startCmd.Flags().String("path", "", "gateway URL path (--http/--sse); default /mcp/<name>")

	syncOpencodeCmd := &cobra.Command{
		Use:   "opencode",
		Short: "Generate OpenCode MCP config from Wuji definitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			http, _ := cmd.Flags().GetBool("http")
			write, _ := cmd.Flags().GetBool("write")

			transport := mcp.TransportStdio
			if http {
				transport = mcp.TransportHTTP
			}

			cfg, err := mcp.BuildOpenCodeConfig(root, mcp.WujiBinary(), transport, parseGatewayFlags(cmd))
			if err != nil {
				return err
			}
			if len(cfg.MCP) == 0 {
				fmt.Fprintln(os.Stdout, "No MCP servers configured.")
				return nil
			}

			if write {
				path, err := mcp.WriteOpenCodeConfig(root, cfg)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Wrote OpenCode MCP config to %s\n", path)
				return nil
			}

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
	syncOpencodeCmd.Flags().Bool("http", false, "generate type: remote entries with gateway URLs (default: local --stdio)")
	syncOpencodeCmd.Flags().Bool("write", false, "write opencode.json in project root")
	syncOpencodeCmd.Flags().String("host", "127.0.0.1", "gateway host for --http")
	syncOpencodeCmd.Flags().Int("port", 0, "gateway port for --http; 0 = auto")
	syncOpencodeCmd.Flags().Int("port-base", 9333, "first auto port for multiple servers")
	syncOpencodeCmd.Flags().String("path", "", "gateway path prefix; default /mcp/<name>")

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync MCP config to external tools",
	}
	syncCmd.AddCommand(syncOpencodeCmd)

	stopCmd := &cobra.Command{
		Use:   "stop [name]",
		Short: "Stop a running MCP server (or all with --all)",
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")

			if all || len(args) == 0 {
				stopped, errs := mgr.StopAll()
				for _, name := range stopped {
					fmt.Fprintf(os.Stdout, "Stopped %s\n", name)
				}
				for _, err := range errs {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				if len(errs) > 0 {
					return fmt.Errorf("%d server(s) failed to stop", len(errs))
				}
				return nil
			}

			if err := mgr.Stop(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Stopped %s\n", args[0])
			return nil
		},
	}
	stopCmd.Flags().Bool("all", false, "stop all running stdio MCP servers")

	restartCmd := &cobra.Command{
		Use:   "restart <name>",
		Short: "Restart a stdio MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := mgr.Restart(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Restarted %s\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(startCmd, stopCmd, restartCmd, syncCmd)
	return cmd
}

func parseGatewayFlags(cmd *cobra.Command) mcp.GatewayOptions {
	host, _ := cmd.Flags().GetString("host")
	path, _ := cmd.Flags().GetString("path")
	port, _ := cmd.Flags().GetInt("port")
	portBase, _ := cmd.Flags().GetInt("port-base")
	return mcp.GatewayOptions{
		Host:     host,
		Port:     port,
		Path:     path,
		PortBase: portBase,
	}
}

func runMCPImport(root, importPath string, overwrite bool) error {
	path, err := wujicfg.ResolveMCPImportPath(importPath)
	if err != nil {
		return err
	}
	n, err := wujicfg.ImportMCP(root, path, overwrite)
	if err != nil {
		return err
	}
	if n == 0 {
		fmt.Fprintln(os.Stdout, "No new servers imported (use --overwrite to replace existing).")
		return nil
	}
	fmt.Fprintf(os.Stdout, "Imported %d server(s) from %s into .wuji/mcp.json\n", n, path)
	return nil
}

func printMCPList(ctx context.Context, mgr *mcp.Manager) error {
	list, err := mgr.ListStatus(ctx)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Println("No MCP servers configured.")
		fmt.Println("Hint: wuji mcp --import          # from Cursor config")
		fmt.Println("      wuji mcp add <name> --command …")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tDETAILS")
	for _, st := range list {
		typ := "stdio"
		status := "stopped"
		details := st.Def.Command
		if st.Def.IsRemote() {
			typ = "remote"
			details = st.Def.URL
			if st.RemoteOK {
				status = "reachable"
			} else {
				status = "unreachable"
			}
		} else if st.Running {
			status = "running"
			if st.GatewayURL != "" {
				details = st.GatewayURL
			} else {
				details = fmt.Sprintf("pid %d", st.PID)
			}
		}
		if st.Def.Disabled {
			status = "disabled"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", st.Name, typ, status, details)
	}
	return w.Flush()
}

func printMCPDetail(st mcp.ServerStatus) error {
	fmt.Printf("Name:     %s\n", st.Name)
	if st.Def.Disabled {
		fmt.Println("Disabled: true")
	}
	if st.Def.IsRemote() {
		fmt.Printf("Type:     remote\n")
		fmt.Printf("URL:      %s\n", st.Def.URL)
		if st.RemoteOK {
			fmt.Println("Status:   reachable")
		} else {
			fmt.Println("Status:   unreachable")
		}
		return nil
	}

	fmt.Printf("Type:     stdio\n")
	fmt.Printf("Command:  %s\n", st.Def.Command)
	if len(st.Def.Args) > 0 {
		fmt.Printf("Args:     %s\n", strings.Join(st.Def.Args, " "))
	}
	if st.Def.Cwd != "" {
		fmt.Printf("Cwd:      %s\n", st.Def.Cwd)
	}
	if len(st.Def.Env) > 0 {
		fmt.Println("Env:")
		for k, v := range st.Def.Env {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}
	if st.Running {
		fmt.Printf("Status:   running (pid %d)\n", st.PID)
		if st.Transport != "" {
			fmt.Printf("Transport: %s\n", st.Transport)
		}
		if st.GatewayURL != "" {
			fmt.Printf("URL:      %s\n", st.GatewayURL)
		}
		if !st.Started.IsZero() {
			fmt.Printf("Started:  %s\n", st.Started.Format(time.RFC3339))
		}
		if st.LogFile != "" {
			fmt.Printf("Log:      %s\n", st.LogFile)
		}
	} else {
		fmt.Println("Status:   stopped")
	}
	return nil
}

func parseEnvPairs(pairs []string) map[string]string {
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok || k == "" {
			continue
		}
		out[k] = v
	}
	return out
}
