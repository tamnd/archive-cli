// Package cli builds the archive command tree on top of the ia library.
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// App carries the resolved configuration and shared clients for a command run.
type App struct {
	Cfg    ia.Config
	HTTP   *ia.HTTPClient
	Cache  *ia.Cache
	Creds  *ia.Credentials
	Out    *Output
	Limit  int
	quiet  bool
	yes    bool
	dryRun bool
}

// globalFlags holds the persistent flag values before they are folded into App.
type globalFlags struct {
	output   string
	fields   string
	limit    int
	dataDir  string
	workers  int
	rate     time.Duration
	retries  int
	timeout  time.Duration
	quiet    bool
	verbose  int
	color    string
	noCache  bool
	yes      bool
	dryRun   bool
	noHeader bool
	template string
	access   string
	secret   string
}

// Root builds the root command and its whole subtree.
func Root() *cobra.Command {
	g := &globalFlags{}
	app := &App{}

	root := &cobra.Command{
		Use:   "archive",
		Short: "A delightful command line for the Internet Archive",
		Long: `archive is the fastest way to work with the Internet Archive from your terminal.

Search millions of items, inspect their metadata, download and verify files,
read view counts, and travel through the Wayback Machine's capture history,
all from one binary with no credentials needed for the public data.

Quick start:
  archive search 'collection:nasa' -n 5      find items
  archive item nasa                          what an item is
  archive files nasa --format JPEG -o url    a file listing
  archive download nasa --format JPEG -d .   download files
  archive wayback get example.com -t 2010    a page as it was in 2010`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return app.init(g)
		},
	}

	pf := root.PersistentFlags()
	def := ia.DefaultConfig()
	pf.StringVarP(&g.output, "output", "o", "auto", "table|json|jsonl|csv|tsv|url|raw")
	pf.StringVar(&g.fields, "fields", "", "comma-separated columns to show")
	pf.IntVarP(&g.limit, "limit", "n", 0, "max results (0 = unlimited)")
	pf.StringVar(&g.dataDir, "data-dir", def.DataDir, "root data directory")
	pf.IntVarP(&g.workers, "workers", "j", def.Workers, "concurrency")
	pf.DurationVar(&g.rate, "rate", def.Delay, "minimum delay between requests")
	pf.IntVar(&g.retries, "retries", def.Retries, "retry attempts on 429/5xx")
	pf.DurationVar(&g.timeout, "timeout", def.Timeout, "per-request timeout")
	pf.BoolVarP(&g.quiet, "quiet", "q", false, "suppress progress output")
	pf.CountVarP(&g.verbose, "verbose", "v", "increase verbosity (repeatable)")
	pf.StringVar(&g.color, "color", "auto", "color output: auto|always|never")
	pf.BoolVar(&g.noCache, "no-cache", false, "bypass on-disk caches")
	pf.BoolVarP(&g.yes, "yes", "y", false, "assume yes to prompts")
	pf.BoolVar(&g.dryRun, "dry-run", false, "print actions without performing them")
	pf.BoolVar(&g.noHeader, "no-header", false, "omit the header row in table/csv output")
	pf.StringVar(&g.template, "template", "", "Go text/template applied per row")
	pf.StringVar(&g.access, "access", "", "IAS3 access key (overrides config/env)")
	pf.StringVar(&g.secret, "secret", "", "IAS3 secret key (overrides config/env)")

	root.AddCommand(
		newSearchCmd(app),
		newItemCmd(app),
		newMetadataCmd(app),
		newFilesCmd(app),
		newDownloadCmd(app),
		newUploadCmd(app),
		newDeleteCmd(app),
		newWaybackCmd(app),
		newViewsCmd(app),
		newTasksCmd(app),
		newOpenCmd(app),
		newConfigureCmd(app),
		newWhoamiCmd(app),
		newConfigCmd(app),
		newCacheCmd(app),
		newVersionCmd(),
	)
	return root
}

func (a *App) init(g *globalFlags) error {
	cfg := ia.DefaultConfig()
	if g.dataDir != "" {
		cfg.DataDir = g.dataDir
		cfg.CacheDir = g.dataDir + "/cache"
	}
	cfg.Workers = max(g.workers, 1)
	cfg.Delay = g.rate
	cfg.Retries = g.retries
	cfg.Timeout = g.timeout

	a.Cfg = cfg
	a.Creds = ia.ResolveCredentials(g.access, g.secret)
	a.HTTP = ia.NewHTTPClient(cfg).WithCredentials(a.Creds)
	a.Cache = ia.NewCache(cfg.CacheDir, !g.noCache)
	a.Limit = g.limit
	a.quiet = g.quiet
	a.yes = g.yes
	a.dryRun = g.dryRun
	a.Out = newOutput(g)
	return nil
}

// requireCreds returns a coded auth error when no credentials are configured.
func (a *App) requireCreds() error {
	if a.Creds.Valid() {
		return nil
	}
	return authErr("no credentials configured; run 'archive configure' or set ARCHIVE_ACCESS_KEY/ARCHIVE_SECRET_KEY")
}

// Execute runs the root command, mapping errors to exit codes.
func Execute(ctx context.Context, cmd *cobra.Command) int {
	if err := cmd.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "archive: "+err.Error())
		if ec, ok := err.(exitCoder); ok {
			return ec.ExitCode()
		}
		return 1
	}
	return 0
}

type exitCoder interface{ ExitCode() int }

// ExitCode maps an error returned from Execute/fang to a process exit code:
// the codedError value when present, 2 for cobra usage errors (unknown command
// or flag), otherwise 1.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if ec, ok := err.(exitCoder); ok {
		return ec.ExitCode()
	}
	msg := err.Error()
	for _, p := range []string{"unknown command", "unknown flag", "unknown shorthand", "invalid argument", "required flag", "accepts ", "requires "} {
		if strings.Contains(msg, p) {
			return 2
		}
	}
	return 1
}

type codedError struct {
	err  error
	code int
}

func (e codedError) Error() string { return e.err.Error() }
func (e codedError) ExitCode() int { return e.code }

func usageErr(msg string) error    { return codedError{fmt.Errorf("%s", msg), 2} }
func noResults(msg string) error   { return codedError{fmt.Errorf("%s", msg), 3} }
func authErr(msg string) error     { return codedError{fmt.Errorf("%s", msg), 4} }
func notFoundErr(msg string) error { return codedError{fmt.Errorf("%s", msg), 5} }

// mapErr converts library sentinel errors to coded CLI errors.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*ia.NotFoundError); ok {
		return notFoundErr(err.Error())
	}
	if err == ia.ErrNoCredentials {
		return authErr(err.Error())
	}
	if ae, ok := err.(*ia.APIError); ok {
		switch ae.Status {
		case 401, 403:
			return authErr(err.Error() + " (authentication required)")
		case 404:
			return notFoundErr(err.Error())
		}
	}
	return err
}
