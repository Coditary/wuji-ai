package cli

import (
	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/core"
	"github.com/coditary/wuji-core/pkg/driver/dummy"
	"github.com/spf13/cobra"
)

func (app *App) resolveDriver(cmd *cobra.Command, cap capability.Type) string {
	preferred := ""
	if cmd != nil {
		d, err := cmd.Root().PersistentFlags().GetString("driver")
		if err == nil {
			preferred = d
		}
	}
	defaultDriver := app.Config.DefaultDriver
	if defaultDriver == "" {
		defaultDriver = dummy.DriverID
	}
	return core.ResolveDriverID(app.Config, defaultDriver, preferred, cap)
}
