package cli

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newSearchCmd(app *App) *cobra.Command {
	var (
		fields     []string
		sorts      []string
		rows       int
		all        bool
		mediatype  string
		collection string
		creator    string
		year       string
		countOnly  bool
	)
	cmd := &cobra.Command{
		Use:     "search <query>",
		Aliases: []string{"find"},
		Short:   "Search the Internet Archive's items",
		Long: `Search items with a Lucene query over indexed metadata.

By default the bounded, sortable Advanced Search endpoint is used. With --all
the cursor-based Scraping API exports every match (no sort, but unbounded), the
right tool for piping millions of identifiers into download.

Examples:
  archive search nasa -n 10
  archive search 'collection:librivoxaudio' --sort 'addeddate desc' -n 20 -o url
  archive search 'mediatype:image' --creator NASA --all -f identifier -o url
  archive search 'identifier:nasa' -f '*' -o json   # every indexed field
  archive search 'collection:prelinger' --count`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			q := buildQuery(strings.Join(args, " "), mediatype, collection, creator, year)
			sq := ia.SearchQuery{Query: q, Fields: fields, Sorts: sorts, Rows: rows}

			if countOnly {
				res, err := ia.Search(c.Context(), app.HTTP, app.Cache, ia.SearchQuery{Query: q, Fields: []string{"identifier"}, Rows: 0, Page: 1})
				if err != nil {
					return mapErr(err)
				}
				if err := app.Out.Emit(Row{Cols: []string{"count"}, Vals: []string{itoa(res.NumFound)}, Value: map[string]any{"count": res.NumFound, "query": q}}); err != nil {
					return err
				}
				return app.Out.Flush()
			}

			app.Out.SetURLField("identifier")
			emit := func(d ia.SearchDoc) error { return app.Out.Emit(searchRow(d, fields)) }

			var n int
			var err error
			if all {
				n, err = ia.Scrape(c.Context(), app.HTTP, sq, app.Limit, emit)
			} else {
				n, err = ia.SearchEach(c.Context(), app.HTTP, app.Cache, sq, app.Limit, emit)
			}
			if err != nil {
				return mapErr(err)
			}
			if err := app.Out.Flush(); err != nil {
				return err
			}
			if n == 0 {
				return noResults("no items matched")
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringSliceVarP(&fields, "field", "f", nil, "metadata field to return (repeatable; '*' = all fields)")
	f.StringSliceVar(&sorts, "sort", nil, "sort key, e.g. 'downloads desc' (repeatable)")
	f.IntVar(&rows, "rows", 0, "page size for Advanced Search")
	f.BoolVar(&all, "all", false, "export every match via the cursor (Scraping API)")
	f.StringVar(&mediatype, "media", "", "filter by mediatype (audio|texts|movies|image|software|data)")
	f.StringVar(&collection, "collection", "", "filter by collection")
	f.StringVar(&creator, "creator", "", "filter by creator")
	f.StringVar(&year, "year", "", "filter by year or range A-B")
	f.BoolVar(&countOnly, "count", false, "print only the number of matches")
	return cmd
}

// buildQuery folds the convenience flags into the Lucene query string.
func buildQuery(base, mediatype, collection, creator, year string) string {
	parts := []string{}
	if b := strings.TrimSpace(base); b != "" {
		parts = append(parts, b)
	}
	if mediatype != "" {
		parts = append(parts, "mediatype:"+mediatype)
	}
	if collection != "" {
		parts = append(parts, "collection:"+collection)
	}
	if creator != "" {
		parts = append(parts, `creator:`+ia.QuoteQuery(creator))
	}
	if year != "" {
		if lo, hi, ok := strings.Cut(year, "-"); ok {
			parts = append(parts, "year:["+lo+" TO "+hi+"]")
		} else {
			parts = append(parts, "year:"+year)
		}
	}
	if len(parts) == 0 {
		return "*:*"
	}
	return strings.Join(parts, " AND ")
}

func searchRow(d ia.SearchDoc, fields []string) Row {
	// A wildcard request (-f '*') fetches every field; the document carries them
	// all for json/jsonl, but the table falls back to the curated columns so it
	// stays readable. --fields can still project any specific one.
	cols := fields
	if len(cols) == 0 || slices.Contains(cols, "*") {
		cols = ia.DefaultFields
	}
	vals := make([]string, len(cols))
	for i, c := range cols {
		vals[i] = d.String(c)
	}
	return Row{Cols: cols, Vals: vals, Value: d}
}
