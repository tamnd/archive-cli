package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newTasksCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks <identifier>",
		Short: "Show the catalog/derive task history of an item",
		Long: `List the catalog tasks (derive, metadata updates, book operations) for an item.
The Archive requires credentials to see tasks for items you do not own; archive
attaches them automatically when configured.

Examples:
  archive tasks nasa
  archive tasks my-item -o jsonl`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			tasks, err := ia.GetTasks(c.Context(), app.HTTP, args[0])
			if err != nil {
				return mapErr(err)
			}
			for _, t := range tasks {
				val := t.Fields()
				if err := app.Out.Emit(Row{
					Cols:  []string{"task_id", "cmd", "status", "server", "submittime"},
					Vals:  []string{fmt.Sprintf("%d", t.TaskID), t.Cmd, t.Status, t.Server, t.DateSub},
					Value: val,
				}); err != nil {
					return err
				}
			}
			if err := app.Out.Flush(); err != nil {
				return err
			}
			if len(tasks) == 0 {
				return noResults("no tasks (or credentials required for this item)")
			}
			return nil
		},
	}
	return cmd
}
