package cli

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newItemCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "item <identifier>",
		Aliases: []string{"show"},
		Short:   "Show a friendly summary of an item",
		Long: `Print the key facts about an item: title, creator, date, mediatype, the
collections it belongs to, its size and file count, and its download/details
URLs. Use 'archive metadata' for the raw JSON or 'archive files' to list files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			m, err := ia.GetMetadata(c.Context(), app.HTTP, app.Cache, id)
			if err != nil {
				return mapErr(err)
			}

			// The field/value summary is for humans; in machine-readable JSON
			// return the full Metadata document so the friendly view never costs
			// information.
			switch app.Out.Format() {
			case FormatJSON:
				var buf bytes.Buffer
				if json.Indent(&buf, m.Raw, "", "  ") == nil {
					return app.Out.Raw(append(buf.Bytes(), '\n'))
				}
				return app.Out.Raw(append(m.Raw, '\n'))
			case FormatJSONL:
				var buf bytes.Buffer
				if json.Compact(&buf, m.Raw) == nil {
					return app.Out.Raw(append(buf.Bytes(), '\n'))
				}
				return app.Out.Raw(append(m.Raw, '\n'))
			}

			pairs := [][2]string{
				{"identifier", id},
				{"title", m.Title()},
				{"creator", strings.Join(m.Meta.Strings("creator"), ", ")},
				{"date", m.Meta.Get("date")},
				{"publicdate", m.Meta.Get("publicdate")},
				{"addeddate", m.Meta.Get("addeddate")},
				{"mediatype", m.Meta.Get("mediatype")},
				{"language", strings.Join(m.Meta.Strings("language"), ", ")},
				{"collection", strings.Join(m.Meta.Strings("collection"), ", ")},
				{"subject", strings.Join(m.Meta.Strings("subject"), ", ")},
				{"licenseurl", m.Meta.Get("licenseurl")},
				{"files", itoa(len(m.Files))},
				{"size", humanBytes(m.ItemSize)},
				{"server", m.Server},
				{"details", ia.DetailsURLFor(id)},
			}
			app.Out.SetURLField("details")
			for _, p := range pairs {
				if p[1] == "" {
					continue
				}
				if err := app.Out.Emit(Row{
					Cols:  []string{"field", "value"},
					Vals:  []string{p[0], p[1]},
					Value: map[string]any{"field": p[0], "value": p[1]},
				}); err != nil {
					return err
				}
			}
			return app.Out.Flush()
		},
	}
	return cmd
}
