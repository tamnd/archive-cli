package ia

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCDXEndToEndPreservesAllColumns drives the CDX parser against a server that
// returns more than the default seven columns, proving every column (urlkey and
// the extra robotflags/redirect) flows through to the record. The live CDX host
// is per-IP rate limited, so this is the reliable coverage for the path.
func TestCDXEndToEndPreservesAllColumns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[
			["urlkey","timestamp","original","mimetype","statuscode","digest","length","robotflags","redirect"],
			["com,example)/","20100101000000","http://example.com/","text/html","200","ABC123","1234","-","-"],
			["com,example)/","20120202000000","http://example.com/","text/html","301","DEF456","99","-","http://example.com/new"]
		]`))
	}))
	defer srv.Close()

	old := WaybackCDXURL
	WaybackCDXURL = srv.URL
	defer func() { WaybackCDXURL = old }()

	var recs []CDXRecord
	n, err := CDX(context.Background(), NewHTTPClient(Config{}), CDXQuery{URL: "example.com"}, func(r CDXRecord) error {
		recs = append(recs, r)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || len(recs) != 2 {
		t.Fatalf("got %d records, want 2", n)
	}

	first := recs[0]
	if first.URLKey != "com,example)/" || first.StatusCode != "200" || first.Digest != "ABC123" {
		t.Errorf("typed fields wrong: %+v", first)
	}
	f := first.Fields()
	for _, k := range []string{"urlkey", "robotflags", "redirect", "mimetype", "length"} {
		if _, ok := f[k]; !ok {
			t.Errorf("Fields() dropped column %q", k)
		}
	}
	// The redirect column carries a real value on the second capture; it must
	// survive verbatim rather than being collapsed away.
	if got := recs[1].Fields()["redirect"]; got != "http://example.com/new" {
		t.Errorf("redirect column = %v, want the redirect target", got)
	}
}
