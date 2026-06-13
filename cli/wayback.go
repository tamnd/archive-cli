package cli

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tamnd/archive-cli/ia"
)

func newWaybackCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wayback",
		Aliases: []string{"wb"},
		Short:   "Work with the Wayback Machine",
		Long: `Travel through the Wayback Machine: find the closest snapshot of a URL, list
its full capture history from the CDX server, fetch a snapshot's content, or
save a fresh capture.`,
	}
	cmd.AddCommand(
		newWaybackGetCmd(app),
		newWaybackListCmd(app),
		newWaybackAvailableCmd(app),
		newWaybackSaveCmd(app),
	)
	return cmd
}

func newWaybackAvailableCmd(app *App) *cobra.Command {
	var ts string
	cmd := &cobra.Command{
		Use:   "available <url>",
		Short: "Show the closest archived snapshot of a URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			snap, ok, err := ia.Available(c.Context(), app.HTTP, app.Cache, args[0], ts)
			if err != nil {
				return mapErr(err)
			}
			if !ok {
				return noResults("no snapshot found for " + args[0])
			}
			app.Out.SetURLField("url")
			if err := app.Out.Emit(Row{
				Cols:  []string{"timestamp", "status", "url"},
				Vals:  []string{snap.Timestamp, snap.Status, snap.URL},
				Value: snap,
			}); err != nil {
				return err
			}
			return app.Out.Flush()
		},
	}
	cmd.Flags().StringVarP(&ts, "timestamp", "t", "", "anchor timestamp (YYYYMMDDhhmmss, partial ok)")
	return cmd
}

func newWaybackListCmd(app *App) *cobra.Command {
	var (
		from, to  string
		matchType string
		filters   []string
		collapse  string
		status    string
		mime      string
	)
	cmd := &cobra.Command{
		Use:     "list <url>",
		Aliases: []string{"cdx"},
		Short:   "List the capture history of a URL (CDX server)",
		Long: `List every capture the Wayback Machine has for a URL. The CDX endpoint is
rate-limited; archive throttles and retries with backoff automatically.

Examples:
  archive wayback list example.com --limit 5
  archive wayback list example.com --from 2010 --to 2012 --status 200
  archive wayback cdx 'example.com/*' --match-type prefix --collapse digest`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if status != "" {
				filters = append(filters, "statuscode:"+status)
			}
			if mime != "" {
				filters = append(filters, "mimetype:"+mime)
			}
			q := ia.CDXQuery{
				URL: args[0], From: from, To: to, MatchType: matchType,
				Filters: filters, Collapse: collapse, Limit: app.Limit,
			}
			app.Out.SetURLField("url")
			n, err := ia.CDX(c.Context(), app.HTTP, q, func(r ia.CDXRecord) error {
				return app.Out.Emit(Row{
					Cols: []string{"timestamp", "url", "mimetype", "status", "digest", "length"},
					Vals: []string{r.Timestamp, r.Original, r.MimeType, r.StatusCode, r.Digest, r.Length},
					Value: map[string]any{
						"timestamp": r.Timestamp, "url": r.Original, "mimetype": r.MimeType,
						"status": r.StatusCode, "digest": r.Digest, "length": r.Length,
						"replay": ia.ReplayURL(r.Timestamp, r.Original, false),
					},
				})
			})
			if err != nil {
				return mapErr(err)
			}
			if err := app.Out.Flush(); err != nil {
				return err
			}
			if n == 0 {
				return noResults("no captures for " + args[0])
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&from, "from", "", "earliest timestamp (YYYYMMDDhhmmss, partial ok)")
	f.StringVar(&to, "to", "", "latest timestamp")
	f.StringVar(&matchType, "match-type", "", "exact|prefix|host|domain")
	f.StringArrayVar(&filters, "filter", nil, "raw CDX filter, e.g. '!mimetype:text/html' (repeatable)")
	f.StringVar(&collapse, "collapse", "", "collapse adjacent rows on a field, e.g. digest")
	f.StringVar(&status, "status", "", "shortcut for filter statuscode:<v>")
	f.StringVar(&mime, "mime", "", "shortcut for filter mimetype:<v>")
	return cmd
}

func newWaybackGetCmd(app *App) *cobra.Command {
	var (
		ts      string
		raw     bool
		asText  bool
		links   bool
		outPath string
	)
	cmd := &cobra.Command{
		Use:   "get <url>",
		Short: "Fetch the content of an archived snapshot",
		Long: `Fetch a snapshot of a URL. By default the closest capture is resolved via the
Availability API; --timestamp picks a specific time (partial ok). --raw fetches
the original archived bytes (no Wayback rewriting), --text extracts readable
text, and --links lists the page's hyperlinks.

Examples:
  archive wayback get example.com -t 2010 --text
  archive wayback get example.com --raw > snapshot.html
  archive wayback get https://example.com --links -o url`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			target := args[0]
			snap, ok, err := ia.Available(c.Context(), app.HTTP, app.Cache, target, ts)
			if err != nil {
				return mapErr(err)
			}
			if !ok {
				return noResults("no snapshot found for " + target)
			}
			// Fetch the raw archived bytes for extraction/raw; the replay variant
			// is reachable from the snapshot timestamp.
			body, err := app.HTTP.FetchBytes(c.Context(), ia.ReplayURL(snap.Timestamp, target, true))
			if err != nil {
				return mapErr(err)
			}
			switch {
			case links:
				app.Out.SetURLField("url")
				for _, l := range ia.ExtractLinks(body) {
					if err := app.Out.Emit(Row{Cols: []string{"url", "text"}, Vals: []string{l.URL, l.Text}, Value: l}); err != nil {
						return err
					}
				}
				return app.Out.Flush()
			case asText:
				return app.Out.Raw([]byte(ia.ExtractText(body) + "\n"))
			default:
				_ = raw // raw is the default fetch above
				if outPath != "" {
					return writeFile(outPath, body)
				}
				return app.Out.Raw(body)
			}
		},
	}
	f := cmd.Flags()
	f.StringVarP(&ts, "timestamp", "t", "", "snapshot timestamp (YYYYMMDDhhmmss, partial ok)")
	f.BoolVar(&raw, "raw", true, "fetch the original archived bytes")
	f.BoolVar(&asText, "text", false, "extract readable text")
	f.BoolVar(&links, "links", false, "list the page's hyperlinks")
	f.StringVar(&outPath, "out", "", "write the snapshot to a file instead of stdout")
	return cmd
}

func newWaybackSaveCmd(app *App) *cobra.Command {
	var (
		outlinks   bool
		screenshot bool
		wait       bool
	)
	cmd := &cobra.Command{
		Use:   "save <url>",
		Short: "Save a fresh capture (Save Page Now)",
		Long: `Trigger a fresh Wayback capture. Anonymously this is a fire-and-forget request;
with --outlinks or --screenshot it uses the authenticated SPN2 API and (with
--wait) polls the job to completion. SPN2 requires IAS3 credentials.

Examples:
  archive wayback save https://example.com/
  archive wayback save https://example.com/ --outlinks --wait`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			target := args[0]
			if app.dryRun {
				app.progressf("would save %s (outlinks=%v screenshot=%v)", target, outlinks, screenshot)
				return nil
			}
			if outlinks || screenshot {
				if err := app.requireCreds(); err != nil {
					return err
				}
				job, err := ia.Save(c.Context(), app.HTTP, target, outlinks, screenshot)
				if err != nil {
					return mapErr(err)
				}
				app.progressf("submitted job %s", job.JobID)
				if wait && job.JobID != "" {
					return app.waitSPN(c.Context(), job.JobID)
				}
				return app.Out.Emit(Row{Cols: []string{"job_id", "status"}, Vals: []string{job.JobID, job.Status}, Value: job})
			}
			url, err := ia.SaveAnonymous(c.Context(), app.HTTP, target)
			if err != nil {
				return mapErr(err)
			}
			app.progressf("saved %s", url)
			return app.Out.Emit(Row{Cols: []string{"url"}, Vals: []string{url}, Value: map[string]any{"url": url}})
		},
	}
	f := cmd.Flags()
	f.BoolVar(&outlinks, "outlinks", false, "also capture outbound links (SPN2, needs auth)")
	f.BoolVar(&screenshot, "screenshot", false, "also capture a screenshot (SPN2, needs auth)")
	f.BoolVar(&wait, "wait", false, "poll the SPN2 job until it completes")
	return cmd
}

func (app *App) waitSPN(ctx context.Context, jobID string) error {
	for {
		job, err := ia.SaveStatus(ctx, app.HTTP, jobID)
		if err != nil {
			return mapErr(err)
		}
		switch job.Status {
		case "success":
			app.progressf("done: %s", ia.ReplayURL(job.Timestamp, job.URL, false))
			return app.Out.Emit(Row{Cols: []string{"job_id", "status", "url"}, Vals: []string{jobID, job.Status, job.URL}, Value: job})
		case "error":
			return mapErr(&statusError{job.Message})
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

type statusError struct{ msg string }

func (e *statusError) Error() string { return "save failed: " + e.msg }

// writeFile saves bytes to a path, used by wayback get --out.
func writeFile(path string, b []byte) error {
	return os.WriteFile(path, b, 0o644)
}
