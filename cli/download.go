package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
	"golang.org/x/sync/errgroup"
)

func newDownloadCmd(app *App) *cobra.Command {
	var (
		glob    string
		format  string
		outDir  string
		flat    bool
		verify  bool
		sources string
	)
	cmd := &cobra.Command{
		Use:     "download <identifier|-> [files...]",
		Aliases: []string{"dl", "get"},
		Short:   "Download files from an item",
		Long: `Download an item's files into a directory. Without explicit file names every
file is fetched; narrow with --glob or --format, or name files directly. Files
already present with a matching md5 are skipped, and --verify checks the md5 of
each download. Pass '-' as the identifier to read identifiers from stdin (one
per line), the natural sink for 'archive search ... -o url'.

Examples:
  archive download principleofrelat00eins --format PDF -d books --verify
  archive download nasa nasa_meta.xml -d .
  archive search 'collection:nasa' --all -f identifier -o url | archive download - --format JPEG -d nasa`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if outDir == "" {
				outDir = app.Cfg.DownloadDir()
			}
			id := args[0]
			names := args[1:]

			run := func(identifier string) error {
				return app.downloadItem(c.Context(), identifier, names, glob, format, outDir, flat, verify)
			}

			if id == "-" {
				n := 0
				if err := readLines(os.Stdin, func(line string) error {
					n++
					return run(line)
				}); err != nil {
					return err
				}
				if n == 0 {
					return usageErr("no identifiers on stdin")
				}
				return nil
			}
			return run(id)
		},
	}
	f := cmd.Flags()
	f.StringVar(&glob, "glob", "", "only files whose name matches this glob")
	f.StringVar(&format, "format", "", "only files whose format contains this substring")
	f.StringVarP(&outDir, "out-dir", "d", "", "destination directory (default ~/data/archive/download)")
	f.BoolVar(&flat, "flat", false, "drop sub-directories from file names")
	f.BoolVar(&verify, "verify", false, "verify each download against its md5")
	f.StringVar(&sources, "sources", "", "(reserved) restrict to original|derivative")
	return cmd
}

func (app *App) downloadItem(ctx context.Context, id string, names []string, glob, format, outDir string, flat, verify bool) error {
	m, err := ia.GetMetadata(ctx, app.HTTP, app.Cache, id)
	if err != nil {
		return mapErr(err)
	}
	var files []ia.FileInfo
	if len(names) > 0 {
		byName := map[string]ia.FileInfo{}
		for _, f := range m.Files {
			byName[f.Name] = f
		}
		for _, n := range names {
			f, ok := byName[n]
			if !ok {
				return notFoundErr(fmt.Sprintf("%s: no such file in %s", n, id))
			}
			files = append(files, f)
		}
	} else {
		files = m.FilterFiles(glob, format)
	}
	if len(files) == 0 {
		return noResults("no files matched in " + id)
	}

	itemDir := outDir
	if outDir == app.Cfg.DownloadDir() {
		itemDir = outDir + "/" + id
	}

	if app.dryRun {
		for _, f := range files {
			app.progressf("would download %s/%s (%s) -> %s", id, f.Name, humanBytes(f.Size()), itemDir)
		}
		return nil
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(app.Cfg.Workers)
	for _, f := range files {
		g.Go(func() error {
			res, err := ia.DownloadFile(gctx, app.HTTP, m, f, itemDir, verify, flat)
			if err != nil {
				return err
			}
			switch {
			case res.Skipped:
				app.progressf("skip %s (md5 match)", f.Name)
			case res.Verified:
				app.progressf("ok   %s (%s, verified)", f.Name, humanBytes(res.Bytes))
			default:
				app.progressf("ok   %s (%s)", f.Name, humanBytes(res.Bytes))
			}
			return nil
		})
	}
	return mapErr(g.Wait())
}
