package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newViewsCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "views <identifier...>",
		Short: "Show view statistics for items",
		Long: `Print all-time, last-30-day, and last-7-day view counts for one or more items
in a single request.

Examples:
  archive views nasa
  archive views nasa principleofrelat00eins -o table`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			views, err := ia.GetViews(c.Context(), app.HTTP, app.Cache, args)
			if err != nil {
				return mapErr(err)
			}
			for _, id := range args {
				v := views[id]
				if err := app.Out.Emit(Row{
					Cols: []string{"identifier", "all_time", "last_30day", "last_7day"},
					Vals: []string{id, itoa64(v.AllTime), itoa64(v.Last30), itoa64(v.Last7)},
					Value: map[string]any{
						"identifier": id, "all_time": v.AllTime,
						"last_30day": v.Last30, "last_7day": v.Last7,
						"have_data": v.HaveData,
					},
				}); err != nil {
					return err
				}
			}
			return app.Out.Flush()
		},
	}
	return cmd
}
