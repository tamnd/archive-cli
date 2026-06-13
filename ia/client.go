package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HTTPClient is a polite, retrying HTTP client for archive.org. It rate-limits
// requests, retries on 429/5xx with linear backoff (honouring Retry-After),
// supports byte-range requests for resumable downloads, and attaches the IAS3
// "LOW access:secret" authorization header when credentials are present.
type HTTPClient struct {
	c         *http.Client
	download  *http.Client // no timeout, for large file bodies
	retries   int
	delay     time.Duration
	userAgent string
	creds     *Credentials // optional; nil means anonymous

	mu   sync.Mutex
	next time.Time // earliest time the next request may start
}

// NewHTTPClient builds an HTTPClient from cfg.
func NewHTTPClient(cfg Config) *HTTPClient {
	ua := cfg.UserAgent
	if ua == "" {
		ua = UserAgent
	}
	return &HTTPClient{
		c:         &http.Client{Timeout: cfg.Timeout},
		download:  &http.Client{},
		retries:   max(cfg.Retries, 0),
		delay:     cfg.Delay,
		userAgent: ua,
	}
}

// WithCredentials attaches credentials used by authenticated requests. Passing
// nil leaves the client anonymous.
func (h *HTTPClient) WithCredentials(c *Credentials) *HTTPClient {
	h.creds = c
	return h
}

// throttle blocks until the configured minimum inter-request delay has elapsed.
func (h *HTTPClient) throttle(ctx context.Context) error {
	if h.delay <= 0 {
		return nil
	}
	h.mu.Lock()
	now := time.Now()
	wait := time.Until(h.next)
	if h.next.Before(now) {
		h.next = now.Add(h.delay)
	} else {
		h.next = h.next.Add(h.delay)
	}
	h.mu.Unlock()
	if wait <= 0 {
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Get fetches url with retries.
func (h *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	return h.do(ctx, h.c, http.MethodGet, url, "", nil, nil)
}

// GetRange fetches the [offset, offset+length) byte span of url. A length <= 0
// requests from offset to the end (used to resume a partial download).
func (h *HTTPClient) GetRange(ctx context.Context, url string, offset, length int64) (*http.Response, error) {
	var rangeHdr string
	if length > 0 {
		rangeHdr = fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)
	} else {
		rangeHdr = fmt.Sprintf("bytes=%d-", offset)
	}
	return h.do(ctx, h.download, http.MethodGet, url, rangeHdr, nil, nil)
}

// GetDownload fetches url with no client timeout (relies on ctx cancellation),
// for large file bodies.
func (h *HTTPClient) GetDownload(ctx context.Context, url string) (*http.Response, error) {
	return h.do(ctx, h.download, http.MethodGet, url, "", nil, nil)
}

// FetchBytes fetches url and returns the whole body, erroring on non-2xx.
func (h *HTTPClient) FetchBytes(ctx context.Context, url string) ([]byte, error) {
	resp, err := h.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{URL: url}
	}
	if resp.StatusCode/100 != 2 {
		return nil, &APIError{Status: resp.StatusCode, URL: url}
	}
	return io.ReadAll(resp.Body)
}

// GetJSON fetches url and decodes the JSON body into v.
func (h *HTTPClient) GetJSON(ctx context.Context, url string, v any) error {
	b, err := h.FetchBytes(ctx, url)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// PostForm posts form-encoded values to url (used by xauthn and SPN2) and
// returns the response body. Credentials, when present, are attached.
func (h *HTTPClient) PostForm(ctx context.Context, rawURL string, form url.Values) ([]byte, error) {
	body := strings.NewReader(form.Encode())
	resp, err := h.do(ctx, h.c, http.MethodPost, rawURL, "", body, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return b, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, rawURL, strings.TrimSpace(string(b)))
	}
	return b, nil
}

// Put uploads body to url with the given extra headers. It requires
// credentials. Returns the response for the caller to inspect.
func (h *HTTPClient) Put(ctx context.Context, url string, body io.Reader, size int64, headers map[string]string) (*http.Response, error) {
	if !h.creds.Valid() {
		return nil, ErrNoCredentials
	}
	resp, err := h.do(ctx, h.download, http.MethodPut, url, "", body, headers, withSize(size))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Delete removes url over IAS3. It requires credentials.
func (h *HTTPClient) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	if !h.creds.Valid() {
		return nil, ErrNoCredentials
	}
	return h.do(ctx, h.c, http.MethodDelete, url, "", nil, headers)
}

type reqOpt func(*http.Request)

func withSize(n int64) reqOpt {
	return func(r *http.Request) {
		if n >= 0 {
			r.ContentLength = n
		}
	}
}

func (h *HTTPClient) do(ctx context.Context, client *http.Client, method, rawURL, rangeHdr string, body io.Reader, headers map[string]string, opts ...reqOpt) (*http.Response, error) {
	// A retriable body must be re-readable; only PUT carries one and the
	// downloader passes a *os.File or *bytes.Reader, which we seek between
	// attempts. For simplicity we only retry idempotent bodyless requests and
	// the first PUT attempt; non-seekable bodies are sent once.
	seeker, _ := body.(io.Seeker)
	var last error
	for i := 0; i <= h.retries; i++ {
		if i > 0 {
			if seeker != nil {
				if _, err := seeker.Seek(0, io.SeekStart); err != nil {
					return nil, err
				}
			} else if body != nil {
				// Non-seekable body cannot be replayed; stop retrying.
				break
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(h.backoff(i)):
			}
		}
		if err := h.throttle(ctx); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", h.userAgent)
		if rangeHdr != "" {
			req.Header.Set("Range", rangeHdr)
		}
		if h.creds.Valid() {
			req.Header.Set("Authorization", h.creds.AuthHeader())
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		for _, o := range opts {
			o(req)
		}
		resp, err := client.Do(req)
		if err != nil {
			last = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			h.honorRetryAfter(resp)
			_ = resp.Body.Close()
			last = fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
			continue
		}
		return resp, nil
	}
	if last == nil {
		last = fmt.Errorf("request failed")
	}
	return nil, fmt.Errorf("all attempts failed for %s: %w", rawURL, last)
}

// backoff returns the wait before retry attempt i (1-based), a gentle quadratic
// ramp over the base delay so the rate-limited CDX host gets room to recover.
func (h *HTTPClient) backoff(i int) time.Duration {
	base := h.delay
	if base <= 0 {
		base = 500 * time.Millisecond
	}
	return base * time.Duration(i*i+1)
}

// honorRetryAfter pushes the next-request gate out by a server-provided
// Retry-After (seconds) so a 429 from the CDX host actually slows us down.
func (h *HTTPClient) honorRetryAfter(resp *http.Response) {
	ra := resp.Header.Get("Retry-After")
	if ra == "" {
		return
	}
	secs, err := strconv.Atoi(strings.TrimSpace(ra))
	if err != nil || secs <= 0 {
		return
	}
	h.mu.Lock()
	until := time.Now().Add(time.Duration(secs) * time.Second)
	if until.After(h.next) {
		h.next = until
	}
	h.mu.Unlock()
}

// NotFoundError marks a 404 so the CLI can map it to exit code 5.
type NotFoundError struct{ URL string }

func (e *NotFoundError) Error() string { return "not found: " + e.URL }

// APIError carries a non-2xx HTTP status so the CLI can map it to an exit code
// (401/403 -> auth required).
type APIError struct {
	Status int
	URL    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP %d from %s", e.Status, e.URL)
}
