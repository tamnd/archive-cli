package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newConfigCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show resolved configuration and data paths",
		Long:  "Print where archive reads and writes, and the effective client settings, so you can see exactly what a run will do.",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the effective configuration",
		RunE: func(c *cobra.Command, _ []string) error {
			cfg := app.Cfg
			rows := [][2]string{
				{"data_dir", cfg.DataDir},
				{"download_dir", cfg.DownloadDir()},
				{"cache_dir", cfg.CacheDir},
				{"config_dir", ia.ConfigDir()},
				{"credentials", ia.CredentialsPath()},
				{"authenticated", boolWord(app.Creds.Valid(), "yes", "no")},
				{"workers", itoa(cfg.Workers)},
				{"rate", cfg.Delay.String()},
				{"timeout", cfg.Timeout.String()},
				{"retries", itoa(cfg.Retries)},
				{"user_agent", cfg.UserAgent},
			}
			for _, r := range rows {
				if err := app.Out.Emit(Row{
					Cols:  []string{"key", "value"},
					Vals:  []string{r[0], r[1]},
					Value: map[string]any{"key": r[0], "value": r[1]},
				}); err != nil {
					return err
				}
			}
			return app.Out.Flush()
		},
	})
	return cmd
}
