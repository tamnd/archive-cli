package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
	"golang.org/x/term"
)

func newConfigureCmd(app *App) *cobra.Command {
	var email string
	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"login"},
		Short:   "Store IAS3 credentials for authenticated commands",
		Long: `Save the IAS3 keys that upload, delete, authenticated Save Page Now, and
own-item task listing need. Two ways:

  - Paste the keys directly with --access and --secret (from
    https://archive.org/account/s3.php).
  - Or log in with your archive.org email and password; archive fetches the
    keys for you. The password is read without echo and never stored.

The keys are written to ~/.config/archive/credentials with mode 0600.

Examples:
  archive configure --access KEY --secret SECRET
  archive configure              # interactive email + password`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			// Path 1: keys provided as flags.
			if app.Creds.Valid() && (cmdFlagChanged(c, "access") || cmdFlagChanged(c, "secret")) {
				return saveCreds(app.Creds)
			}

			// Path 2: interactive login.
			reader := bufio.NewReader(os.Stdin)
			if email == "" {
				_, _ = fmt.Fprint(cmdErr, "Email: ")
				line, _ := reader.ReadString('\n')
				email = strings.TrimSpace(line)
			}
			if email == "" {
				return usageErr("email is required (or pass --access/--secret)")
			}
			_, _ = fmt.Fprint(cmdErr, "Password: ")
			pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			_, _ = fmt.Fprintln(cmdErr)
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}
			creds, err := ia.Login(c.Context(), app.HTTP, email, string(pwBytes))
			if err != nil {
				return authErr(err.Error())
			}
			return saveCreds(creds)
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "archive.org account email (for password login)")
	return cmd
}

func saveCreds(c *ia.Credentials) error {
	if err := c.Save(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmdErr, "saved credentials to %s\n", ia.CredentialsPath())
	return nil
}

func cmdFlagChanged(c *cobra.Command, name string) bool {
	f := c.Flags().Lookup(name)
	return f != nil && f.Changed
}
