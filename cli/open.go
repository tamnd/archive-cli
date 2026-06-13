package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newOpenCmd(app *App) *cobra.Command {
	var web bool
	cmd := &cobra.Command{
		Use:   "open <identifier|url>",
		Short: "Print (or open) the details / Wayback URL for a thing",
		Long: `Print the canonical archive.org URL for an item identifier, or the Wayback
landing page for a web URL. With --web, open it in the default browser.

Examples:
  archive open nasa
  archive open https://example.com --web`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			arg := args[0]
			var url string
			if strings.Contains(arg, "://") || strings.Contains(arg, ".") && !ia.ValidIdentifier(arg) {
				url = "https://web.archive.org/web/2/" + arg
			} else {
				url = ia.DetailsURLFor(arg)
			}
			if web {
				if err := openBrowser(url); err != nil {
					app.progressf("could not open browser: %v", err)
				}
			}
			_, _ = fmt.Fprintln(cmdOut, url)
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "open the URL in the default browser")
	return cmd
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, append(args, url)...).Start()
}
