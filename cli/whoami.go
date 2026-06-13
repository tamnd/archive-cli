package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newWhoamiCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the configured credentials",
		Long:  "Print the resolved IAS3 access key (the secret is masked) and where it came from. Exits 4 when no credentials are configured.",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if !app.Creds.Valid() {
				return app.requireCreds()
			}
			rows := [][2]string{
				{"access", app.Creds.Access},
				{"secret", app.Creds.MaskedSecret()},
				{"source", ia.CredentialsPath()},
			}
			for _, r := range rows {
				if err := app.Out.Emit(Row{
					Cols:  []string{"field", "value"},
					Vals:  []string{r[0], r[1]},
					Value: map[string]any{"field": r[0], "value": r[1]},
				}); err != nil {
					return err
				}
			}
			return app.Out.Flush()
		},
	}
	return cmd
}
