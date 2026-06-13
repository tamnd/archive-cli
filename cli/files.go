package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newFilesCmd(app *App) *cobra.Command {
	var (
		glob   string
		format string
	)
	cmd := &cobra.Command{
		Use:     "files <identifier>",
		Aliases: []string{"ls"},
		Short:   "List the files in an item",
		Long: `List an item's files with their size, format, and md5. Filter by a name glob
(--glob '*.pdf') or a format substring (--format PDF). With -o url the canonical
download URLs are printed, ready to pipe into a downloader.

Examples:
  archive files principleofrelat00eins
  archive files principleofrelat00eins --format PDF -o url
  archive files nasa --glob '*.jpg' -o jsonl`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			m, err := ia.GetMetadata(c.Context(), app.HTTP, app.Cache, id)
			if err != nil {
				return mapErr(err)
			}
			files := m.FilterFiles(glob, format)
			if app.Limit > 0 && len(files) > app.Limit {
				files = files[:app.Limit]
			}
			app.Out.SetURLField("url")
			for _, f := range files {
				url := ia.DownloadURLFor(id, f.Name)
				// Value carries every field the API returned for this file (so
				// json/jsonl/template lose nothing); Cols is the curated table
				// view, and --fields can reach any raw field via the Value map.
				val := f.Fields()
				val["url"] = url
				if err := app.Out.Emit(Row{
					Cols:  []string{"name", "size", "format", "md5", "url"},
					Vals:  []string{f.Name, itoa64(f.Size()), f.Format, f.MD5, url},
					Value: val,
				}); err != nil {
					return err
				}
			}
			if err := app.Out.Flush(); err != nil {
				return err
			}
			if len(files) == 0 {
				return noResults("no files matched")
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&glob, "glob", "", "filter file names by glob (e.g. '*.pdf')")
	f.StringVar(&format, "format", "", "filter by format substring (e.g. PDF)")
	return cmd
}
