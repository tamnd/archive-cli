package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newUploadCmd(app *App) *cobra.Command {
	var (
		meta        []string
		remoteName  string
		makeBucket  bool
		noDerive    bool
		contentType string
	)
	cmd := &cobra.Command{
		Use:     "upload <identifier> <file...>",
		Aliases: []string{"up", "put"},
		Short:   "Upload files into an item over the S3-like interface",
		Long: `Upload one or more local files into an item. Requires IAS3 credentials
(run 'archive configure'). Set metadata with -m key:value (repeatable); these
apply when the item is created. --make-bucket creates the item if it does not
exist yet.

Examples:
  archive upload my-item report.pdf -m 'title:My Report' -m mediatype:texts --make-bucket
  archive upload my-item *.jpg
  archive upload my-item big.iso --dry-run`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			id, files := args[0], args[1:]
			if !ia.ValidIdentifier(id) {
				return usageErr("invalid identifier: " + id)
			}
			metadata, err := parseKV(meta)
			if err != nil {
				return usageErr(err.Error())
			}
			if len(files) > 1 && remoteName != "" {
				return usageErr("--name can only be used with a single file")
			}
			opts := ia.UploadOptions{
				Metadata: metadata, RemoteName: remoteName,
				MakeBucket: makeBucket, NoDerive: noDerive, ContentType: contentType,
			}

			if app.dryRun {
				for _, fpath := range files {
					info, err := os.Stat(fpath)
					if err != nil {
						return err
					}
					name := opts.RemoteName
					if name == "" {
						name = baseName(fpath)
					}
					app.progressf("PUT %s", ia.UploadURL(id, name))
					printHeaders(ia.UploadHeaders(opts, info.Size()))
				}
				return nil
			}

			if err := app.requireCreds(); err != nil {
				return err
			}
			for _, fpath := range files {
				url, err := ia.Upload(c.Context(), app.HTTP, id, fpath, opts)
				if err != nil {
					return mapErr(err)
				}
				app.progressf("uploaded %s", url)
				if err := app.Out.Emit(Row{Cols: []string{"file", "url"}, Vals: []string{fpath, url}, Value: map[string]any{"file": fpath, "url": url}}); err != nil {
					return err
				}
			}
			return app.Out.Flush()
		},
	}
	f := cmd.Flags()
	f.StringArrayVarP(&meta, "metadata", "m", nil, "metadata key:value to set on the item (repeatable)")
	f.StringVar(&remoteName, "name", "", "remote file name (single-file uploads only)")
	f.BoolVar(&makeBucket, "make-bucket", false, "create the item if it does not exist")
	f.BoolVar(&noDerive, "no-derive", false, "skip derivation after upload")
	f.StringVar(&contentType, "content-type", "", "override the Content-Type header")
	return cmd
}

func printHeaders(h map[string]string) {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = fmt.Fprintf(cmdErr, "  %s: %s\n", k, h[k])
	}
}

func baseName(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}
