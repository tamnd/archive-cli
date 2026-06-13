package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const availTTL = 10 * time.Minute

// Available returns the closest Wayback snapshot for a URL, optionally anchored
// to a timestamp (YYYYMMDDhhmmss, partial allowed).
func Available(ctx context.Context, h *HTTPClient, cache *Cache, target, timestamp string) (Snapshot, bool, error) {
	// The availability API is order-sensitive: "url" must precede "timestamp"
	// or it returns an empty result, so the query is built by hand rather than
	// via url.Values.Encode (which sorts keys alphabetically).
	u := WaybackAvailURL + "?url=" + url.QueryEscape(target)
	if timestamp != "" {
		u += "&timestamp=" + url.QueryEscape(timestamp)
	}
	var body []byte
	if cache != nil {
		if b, ok := cache.Get(u, availTTL); ok {
			body = b
		}
	}
	if body == nil {
		b, err := h.FetchBytes(ctx, u)
		if err != nil {
			return Snapshot{}, false, err
		}
		body = b
		if cache != nil {
			cache.Put(u, body)
		}
	}
	var r struct {
		ArchivedSnapshots struct {
			Closest *Snapshot `json:"closest"`
		} `json:"archived_snapshots"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return Snapshot{}, false, err
	}
	if r.ArchivedSnapshots.Closest == nil {
		return Snapshot{}, false, nil
	}
	return *r.ArchivedSnapshots.Closest, true, nil
}

// CDXQuery describes a Wayback CDX history request.
type CDXQuery struct {
	URL       string
	From      string
	To        string
	MatchType string   // exact | prefix | host | domain
	Filters   []string // e.g. "statuscode:200", "!mimetype:text/html"
	Collapse  string   // e.g. "digest" or "timestamp:8"
	Limit     int      // negative = newest N
}

// CDX fetches the capture history for a URL and calls fn for each record. The
// CDX server returns a header row first, which is consumed to map columns.
func CDX(ctx context.Context, h *HTTPClient, q CDXQuery, fn func(CDXRecord) error) (int, error) {
	v := url.Values{}
	v.Set("url", q.URL)
	v.Set("output", "json")
	if q.From != "" {
		v.Set("from", q.From)
	}
	if q.To != "" {
		v.Set("to", q.To)
	}
	if q.MatchType != "" {
		v.Set("matchType", q.MatchType)
	}
	for _, f := range q.Filters {
		v.Add("filter", f)
	}
	if q.Collapse != "" {
		v.Set("collapse", q.Collapse)
	}
	if q.Limit != 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	u := WaybackCDXURL + "?" + v.Encode()
	b, err := h.FetchBytes(ctx, u)
	if err != nil {
		return 0, err
	}
	var rows [][]string
	if err := json.Unmarshal(b, &rows); err != nil {
		return 0, fmt.Errorf("decoding CDX response: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	header := rows[0]
	idx := map[string]int{}
	for i, c := range header {
		idx[c] = i
	}
	get := func(row []string, col string) string {
		if i, ok := idx[col]; ok && i < len(row) {
			return row[i]
		}
		return ""
	}
	n := 0
	for _, row := range rows[1:] {
		// Capture every column under its header name so nothing the server
		// returned is lost, then fill the typed convenience fields from it.
		all := make(map[string]string, len(header))
		for i, c := range header {
			if i < len(row) {
				all[c] = row[i]
			}
		}
		rec := CDXRecord{
			URLKey:     get(row, "urlkey"),
			Timestamp:  get(row, "timestamp"),
			Original:   get(row, "original"),
			MimeType:   get(row, "mimetype"),
			StatusCode: get(row, "statuscode"),
			Digest:     get(row, "digest"),
			Length:     get(row, "length"),
			All:        all,
		}
		if err := fn(rec); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// ReplayURL builds a Wayback replay URL. When raw is set it requests the
// original archived bytes (the "id_" modifier) instead of the rewritten page.
func ReplayURL(timestamp, target string, raw bool) string {
	if timestamp == "" {
		timestamp = "2"
	}
	mod := "/"
	if raw {
		mod = "id_/"
	}
	return WaybackReplay + timestamp + mod + target
}

// SPNJob is the result of a Save Page Now request. The SPN2 status reply carries
// more than the typed fields (resources, outlinks, counters, original_url,
// duration_sec, http_status, ...); the full record is kept in Raw so none of it
// is dropped in json/jsonl/template output.
type SPNJob struct {
	JobID     string `json:"job_id"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Message   string `json:"message"`

	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON fills the typed fields and retains the complete raw record.
func (j *SPNJob) UnmarshalJSON(b []byte) error {
	type alias SPNJob // sheds UnmarshalJSON to avoid recursion
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*j = SPNJob(a)
	j.Raw = append(j.Raw[:0], b...)
	return nil
}

// Fields decodes the complete job record so every SPN2 field survives.
func (j SPNJob) Fields() map[string]any {
	out := map[string]any{}
	if len(j.Raw) > 0 {
		_ = json.Unmarshal(j.Raw, &out)
	}
	return out
}

// SaveAnonymous triggers an anonymous Save Page Now capture (fire and forget).
// It returns the replay URL the capture will live at.
func SaveAnonymous(ctx context.Context, h *HTTPClient, target string) (string, error) {
	resp, err := h.Get(ctx, WaybackSaveURL+target)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 && resp.StatusCode != 302 {
		return "", fmt.Errorf("save failed: HTTP %d", resp.StatusCode)
	}
	// The Content-Location header carries the snapshot path when present.
	if loc := resp.Header.Get("Content-Location"); loc != "" {
		return "https://web.archive.org" + loc, nil
	}
	return WaybackSaveURL + target, nil
}

// Save submits an authenticated SPN2 capture and returns the job. Requires
// credentials. capture_outlinks/screenshot follow the SPN2 contract.
func Save(ctx context.Context, h *HTTPClient, target string, outlinks, screenshot bool) (SPNJob, error) {
	if !h.creds.Valid() {
		return SPNJob{}, ErrNoCredentials
	}
	form := url.Values{"url": {target}}
	if outlinks {
		form.Set("capture_outlinks", "1")
	}
	if screenshot {
		form.Set("capture_screenshot", "1")
	}
	body, err := h.PostForm(ctx, "https://web.archive.org/save", form)
	if err != nil {
		return SPNJob{}, err
	}
	var job SPNJob
	if err := json.Unmarshal(body, &job); err != nil {
		return SPNJob{}, fmt.Errorf("unexpected save response: %s", strings.TrimSpace(string(body)))
	}
	return job, nil
}

// SaveStatus polls an SPN2 job once.
func SaveStatus(ctx context.Context, h *HTTPClient, jobID string) (SPNJob, error) {
	var job SPNJob
	err := h.GetJSON(ctx, "https://web.archive.org/save/status/"+jobID, &job)
	return job, err
}
