package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/auth"
	"github.com/coditary/wuji-core/pkg/catalog"
)

var providerHints = map[string]string{
	"anthropic":      "Create an API key at https://console.anthropic.com/",
	"openai":         "Use an OpenAI API key from https://platform.openai.com/api-keys",
	"github-copilot": "Use a GitHub token with Copilot scope, or run: gh auth token",
	"google":         "Use a Google AI API key from https://aistudio.google.com/apikey",
	"openrouter":     "Create an API key at https://openrouter.ai/keys",
	"deepseek":       "Create an API key at https://platform.deepseek.com/",
	"vercel":         "Create an API key at https://vercel.link/ai-gateway-token",
	"amazon-bedrock": "Configure AWS credentials (AWS_PROFILE, AWS_REGION) or a bearer token.",
	"cloudflare-ai-gateway": "Set CLOUDFLARE_GATEWAY_ID, CLOUDFLARE_ACCOUNT_ID, and CLOUDFLARE_API_TOKEN.",
}

func newProviderCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "provider",
		Aliases: []string{"providers", "auth"},
		Short:   "Manage provider credentials (OpenCode-style login)",
		Long: `Store API credentials and configure routes from the models.dev catalog.

  wuji provider login [provider-id]   add credential + apis: route
  wuji provider list                  show saved credentials
  wuji provider logout [provider-id]  remove credential`,
	}

	loginCmd := &cobra.Command{
		Use:     "login [provider-id]",
		Aliases: []string{"connect"},
		Short:   "Log in to a catalog provider and save credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProviderLogin(app, args)
		},
	}

	logoutCmd := &cobra.Command{
		Use:     "logout [provider-id]",
		Aliases: []string{"disconnect"},
		Short:   "Remove a saved provider credential",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProviderLogout(app, args)
		},
	}

	cmd.AddCommand(
		loginCmd,
		logoutCmd,
		&cobra.Command{
			Use:     "list",
			Aliases: []string{"ls"},
			Short:   "List saved credentials and active environment variables",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runProviderList(app)
			},
		},
	)
	return cmd
}

func runProviderLogin(app *App, args []string) error {
	idx, err := loadCatalogIndex(app)
	if err != nil {
		return err
	}

	store, err := auth.Open("")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Add credential")
	fmt.Fprintf(os.Stdout, "Store: %s\n", displayPath(store.Path()))

	providerID, err := resolveProviderSelection(idx, args)
	if err != nil {
		return err
	}

	if providerID == "__other__" {
		id, err := readLine("Enter provider id (a-z, 0-9, hyphens): ", os.Stdin)
		if err != nil {
			return err
		}
		if !isValidProviderID(id) {
			return fmt.Errorf("invalid provider id %q", id)
		}
		providerID = id
		if _, ok := idx.Providers[providerID]; !ok {
			fmt.Fprintf(os.Stdout, "\nNote: %q is not in the catalog — only the credential will be stored.\n", providerID)
			fmt.Fprintln(os.Stdout, "Add an apis: route manually in .wuji/config.yaml if needed.")
		}
	}

	entry, inCatalog := idx.Providers[providerID]
	if hint := providerHints[providerID]; hint != "" {
		fmt.Fprintf(os.Stdout, "\n%s\n", hint)
	}

	cred, err := promptProviderCredential(providerID, entry, inCatalog)
	if err != nil {
		return err
	}
	if err := store.Set(providerID, cred); err != nil {
		return err
	}

	if inCatalog {
		catalog.EnsureAPIRoute(app.Config, providerID, entry)
		if err := app.Config.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}

	fmt.Fprintf(os.Stdout, "\nLogged in to %q\n", providerID)
	if inCatalog {
		model := catalog.APIEntryFromProvider(entry).Text
		if model != nil && model.Defaults != nil && model.Defaults.Model != "" {
			fmt.Fprintf(os.Stdout, "Route configured: apis.%s (default model: %s)\n", providerID, model.Defaults.Model)
		} else {
			fmt.Fprintf(os.Stdout, "Route configured: apis.%s\n", providerID)
		}
	}
	fmt.Fprintln(os.Stdout, "Done")
	return nil
}

func resolveProviderSelection(idx *catalog.Index, args []string) (string, error) {
	if len(args) > 0 {
		input := strings.TrimSpace(args[0])
		if p, ok := idx.Providers[input]; ok {
			return p.ID, nil
		}
		for id, p := range idx.Providers {
			if strings.EqualFold(p.Name, input) {
				return id, nil
			}
		}
		return "", fmt.Errorf("unknown provider %q — run: wuji catalog sync && wuji list providers", input)
	}

	ids := make([]string, 0, len(idx.Providers))
	for id := range idx.Providers {
		ids = append(ids, id)
	}
	priority := map[string]int{
		"openai": 1, "github-copilot": 2, "google": 3, "anthropic": 4, "openrouter": 5,
	}
	sort.Slice(ids, func(i, j int) bool {
		pi, pj := priority[ids[i]], priority[ids[j]]
		if pi != 0 || pj != 0 {
			if pi == 0 {
				return false
			}
			if pj == 0 {
				return true
			}
			if pi != pj {
				return pi < pj
			}
		}
		ni := idx.Providers[ids[i]].Name
		nj := idx.Providers[ids[j]].Name
		if ni == "" {
			ni = ids[i]
		}
		if nj == "" {
			nj = ids[j]
		}
		return ni < nj
	})
	names := make([]string, len(ids))
	for i, id := range ids {
		name := idx.Providers[id].Name
		if name == "" {
			name = id
		}
		names[i] = name
	}

	choice, err := selectOption("Select provider", names)
	if err != nil {
		return "", err
	}
	if choice == len(names) {
		return "__other__", nil
	}
	return ids[choice], nil
}

func promptProviderCredential(providerID string, entry catalog.ProviderEntry, inCatalog bool) (auth.Entry, error) {
	envVars := entry.Env
	if !inCatalog || len(envVars) == 0 {
		key, err := readPassword("Enter your API key: ")
		if err != nil {
			return auth.Entry{}, err
		}
		if key == "" {
			return auth.Entry{}, fmt.Errorf("API key is required")
		}
		return auth.Entry{Type: "api", Key: key}, nil
	}

	if len(envVars) == 1 {
		key, err := readPassword(fmt.Sprintf("Enter your API key (%s): ", envVars[0]))
		if err != nil {
			return auth.Entry{}, err
		}
		if key == "" {
			return auth.Entry{}, fmt.Errorf("credential is required")
		}
		return auth.Entry{
			Type: "api",
			Key:  key,
			Env:  map[string]string{envVars[0]: key},
		}, nil
	}

	values := map[string]string{}
	for _, env := range envVars {
		val, err := readPassword(fmt.Sprintf("Enter value for %s: ", env))
		if err != nil {
			return auth.Entry{}, err
		}
		if val == "" {
			return auth.Entry{}, fmt.Errorf("%s is required", env)
		}
		values[env] = val
	}
	return auth.Entry{Type: "api", Env: values, Key: values[envVars[0]]}, nil
}

func runProviderLogout(app *App, args []string) error {
	store, err := auth.Open("")
	if err != nil {
		return err
	}
	all, err := store.All()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Fprintln(os.Stdout, "No credentials found.")
		return nil
	}

	providerID := ""
	if len(args) > 0 {
		providerID = strings.TrimSpace(args[0])
	} else {
		idx, _ := loadCatalogIndex(app)
		ids := sortedKeys(all)
		labels := make([]string, len(ids))
		for i, id := range ids {
			name := id
			if idx != nil {
				if p, ok := idx.Providers[id]; ok && p.Name != "" {
					name = fmt.Sprintf("%s (%s)", p.Name, all[id].Type)
				}
			}
			labels[i] = name
		}
		choice, err := selectOption("Remove credential", labels)
		if err != nil {
			return err
		}
		providerID = ids[choice]
	}

	if _, ok := all[providerID]; !ok {
		return fmt.Errorf("credential for %q not found", providerID)
	}
	if err := store.Remove(providerID); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Removed credential for %q\n", providerID)
	return nil
}

func runProviderList(app *App) error {
	store, err := auth.Open("")
	if err != nil {
		return err
	}
	all, err := store.All()
	if err != nil {
		return err
	}

	idx, _ := catalog.LoadIndex(app.Config.Root)

	fmt.Fprintf(os.Stdout, "Credentials %s\n\n", displayPath(store.Path()))
	if len(all) == 0 {
		fmt.Fprintln(os.Stdout, "  (none — run: wuji provider login)")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PROVIDER\tTYPE\tNAME")
		for _, id := range sortedKeys(all) {
			name := id
			if idx != nil {
				if p, ok := idx.Providers[id]; ok && p.Name != "" {
					name = p.Name
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", id, all[id].Type, name)
		}
		_ = w.Flush()
	}

	if idx != nil {
		type envActive struct{ provider, env string }
		var active []envActive
		for _, p := range idx.Providers {
			name := p.Name
			if name == "" {
				name = p.ID
			}
			for _, env := range p.Env {
				if strings.TrimSpace(os.Getenv(env)) != "" {
					active = append(active, envActive{name, env})
				}
			}
		}
		if len(active) > 0 {
			fmt.Fprintln(os.Stdout, "\nEnvironment")
			sort.Slice(active, func(i, j int) bool {
				return active[i].provider < active[j].provider
			})
			for _, item := range active {
				fmt.Fprintf(os.Stdout, "  %s  %s\n", item.provider, item.env)
			}
		}
	}
	return nil
}

func sortedKeys(all map[string]auth.Entry) []string {
	ids := make([]string, 0, len(all))
	for id := range all {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func isValidProviderID(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func displayPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return filepath.Join("~", strings.TrimPrefix(path, home))
	}
	return path
}
