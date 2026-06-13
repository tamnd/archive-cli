package cli

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newMetadataCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "metadata <identifier> [subpath]",
		Aliases: []string{"meta"},
		Short:   "Print the raw Metadata API document for an item",
		Long: `Fetch the Metadata API JSON for an item, or a sub-resource of it.

Sub-resources:
  metadata    just the metadata dictionary
  files       the file manifest ({"result":[...]})
  server      the hosting datanode

Examples:
  archive metadata nasa
  archive metadata nasa files
  archive metadata nasa metadata | jq .title`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			sub := ""
			if len(args) == 2 {
				sub = args[1]
			}
			body, err := ia.GetMetadataSub(c.Context(), app.HTTP, id, sub)
			if err != nil {
				return mapErr(err)
			}
			if bytes.Equal(bytes.TrimSpace(body), []byte("{}")) {
				return notFoundErr("no such item: " + id)
			}
			var pretty bytes.Buffer
			if json.Indent(&pretty, body, "", "  ") == nil {
				return app.Out.Raw(append(pretty.Bytes(), '\n'))
			}
			return app.Out.Raw(append(body, '\n'))
		},
	}
	return cmd
}
