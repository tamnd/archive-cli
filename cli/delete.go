package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newDeleteCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <identifier> <remote-file...>",
		Aliases: []string{"rm"},
		Short:   "Delete files from an item over the S3-like interface",
		Long: `Delete one or more files from an item. Requires IAS3 credentials and confirms
unless --yes is given. This removes the file and its derivatives.

Examples:
  archive delete my-item stale.pdf --yes
  archive delete my-item old.jpg --dry-run`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			id, names := args[0], args[1:]
			if app.dryRun {
				for _, n := range names {
					app.progressf("DELETE %s", ia.UploadURL(id, n))
				}
				return nil
			}
			if err := app.requireCreds(); err != nil {
				return err
			}
			if !confirm(app.yes, "delete "+itoa(len(names))+" file(s) from "+id+"?") {
				return usageErr("cancelled")
			}
			for _, n := range names {
				if err := ia.DeleteFile(c.Context(), app.HTTP, id, n); err != nil {
					return mapErr(err)
				}
				app.progressf("deleted %s/%s", id, n)
			}
			return nil
		},
	}
	return cmd
}
