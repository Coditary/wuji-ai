package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/catalog"
)

func newCatalogCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Sync and inspect the models.dev provider/model catalog",
		Long: `Wuji caches a slim index of models.dev provider and model metadata under .wuji/catalog/.

Run sync once (or periodically) to refresh:
  wuji catalog sync

Browse cached data:
  wuji list providers
  wuji list models
  wuji list labs`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "sync",
			Short: "Download catalog.json from models.dev and build the slim index",
			RunE: func(cmd *cobra.Command, args []string) error {
				source := app.Config.ResolvedCatalog().Source
				res, err := catalog.Sync(cmd.Context(), app.Config.Root, source)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "%s\n", res.Message)
				fmt.Fprintf(os.Stdout, "  source:          %s\n", res.Source)
				fmt.Fprintf(os.Stdout, "  synced:          %s\n", res.SyncedAt.UTC().Format(time.RFC3339))
				fmt.Fprintf(os.Stdout, "  index size:      %d KB\n", (res.IndexBytes+1023)/1024)
				fmt.Fprintf(os.Stdout, "  provider models: %d\n", res.ProviderModels)
				fmt.Fprintf(os.Stdout, "  cache:           %s\n", catalog.Dir(app.Config.Root))
				return nil
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show catalog cache age and counts",
			RunE: func(cmd *cobra.Command, args []string) error {
				meta, err := catalog.LoadMeta(app.Config.Root)
				if err != nil {
					return err
				}
				if meta == nil {
					fmt.Println("Catalog not synced yet.")
					fmt.Println("Run: wuji catalog sync")
					return nil
				}
				age, err := catalog.IndexAge(app.Config.Root)
				if err != nil {
					return err
				}
				fmt.Printf("Source:           %s\n", meta.Source)
				fmt.Printf("Synced:           %s\n", meta.SyncedAt)
				fmt.Printf("Age:              %s\n", age.Truncate(time.Second))
				fmt.Printf("Providers:        %d\n", meta.ProviderCount)
				fmt.Printf("Canonical models: %d\n", meta.ModelCount)
				fmt.Printf("Labs:             %d\n", meta.LabCount)
				fmt.Printf("Provider models:  %d\n", meta.ProviderModels)
				fmt.Printf("Cache:            %s\n", catalog.Dir(app.Config.Root))
				return nil
			},
		},
	)
	return cmd
}

func loadCatalogIndex(app *App) (*catalog.Index, error) {
	return catalog.LoadOrSync(context.Background(), app.Config)
}

func printCatalogProviderList(app *App) error {
	idx, err := loadCatalogIndex(app)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tMODELS\tAPI")
	ids := make([]string, 0, len(idx.Providers))
	for id := range idx.Providers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		p := idx.Providers[id]
		api := p.API
		if api == "" {
			api = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", p.ID, p.Name, len(p.Models), api)
	}
	return w.Flush()
}

func printCatalogModelList(app *App, providerID, labID string) error {
	idx, err := loadCatalogIndex(app)
	if err != nil {
		return err
	}

	providerID = strings.TrimSpace(providerID)
	labID = strings.TrimSpace(labID)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if providerID != "" {
		p, ok := idx.Providers[providerID]
		if !ok {
			return fmt.Errorf("provider %q not found in catalog — run: wuji catalog sync", providerID)
		}
		fmt.Fprintf(os.Stdout, "Provider: %s (%s)\n\n", p.Name, p.ID)
		fmt.Fprintln(w, "MODEL ID\tNAME")
		modelIDs := make([]string, 0, len(p.Models))
		for id := range p.Models {
			modelIDs = append(modelIDs, id)
		}
		sort.Strings(modelIDs)
		for _, id := range modelIDs {
			fmt.Fprintf(w, "%s\t%s\n", id, p.Models[id])
		}
		return w.Flush()
	}

	fmt.Fprintln(w, "ID\tNAME\tLAB\tCTX\tREASONING\tTOOLS")
	ids := make([]string, 0, len(idx.Models))
	for id := range idx.Models {
		m := idx.Models[id]
		if labID != "" && m.Lab != labID {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		fmt.Println("No models matched.")
		return nil
	}
	for _, id := range ids {
		m := idx.Models[id]
		ctx := "-"
		if m.Context > 0 {
			ctx = fmt.Sprintf("%d", m.Context)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%t\n", m.ID, m.Name, m.Lab, ctx, m.Reasoning, m.ToolCall)
	}
	return w.Flush()
}

func printCatalogLabList(app *App) error {
	idx, err := loadCatalogIndex(app)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tMODELS")
	ids := make([]string, 0, len(idx.Labs))
	for id := range idx.Labs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		lab := idx.Labs[id]
		fmt.Fprintf(w, "%s\t%s\t%d\n", lab.ID, lab.Name, len(lab.Models))
	}
	return w.Flush()
}
