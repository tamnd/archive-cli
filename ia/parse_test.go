package ia

import (
	"encoding/json"
	"testing"
)

func TestFileInfoCoercion(t *testing.T) {
	var f FileInfo
	if err := json.Unmarshal([]byte(`{"name":"a.pdf","size":"1024","mtime":"1657988342","md5":"abc","format":"PDF"}`), &f); err != nil {
		t.Fatal(err)
	}
	if f.Size() != 1024 {
		t.Errorf("Size = %d, want 1024", f.Size())
	}
	if f.Mtime() != 1657988342 {
		t.Errorf("Mtime = %d, want 1657988342", f.Mtime())
	}
}

func TestMetaDictSingleAndArray(t *testing.T) {
	var d MetaDict
	if err := json.Unmarshal([]byte(`{"title":"One","collection":["a","b"],"year":2009}`), &d); err != nil {
		t.Fatal(err)
	}
	if got := d.Get("title"); got != "One" {
		t.Errorf("title = %q", got)
	}
	if got := d.Strings("collection"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("collection = %v", got)
	}
	if got := d.Get("year"); got != "2009" {
		t.Errorf("year = %q", got)
	}
}

func TestMetadataExists(t *testing.T) {
	var empty Metadata
	if err := json.Unmarshal([]byte(`{}`), &empty); err != nil {
		t.Fatal(err)
	}
	if empty.Exists() {
		t.Error("empty metadata should not Exist")
	}
	var real Metadata
	_ = json.Unmarshal([]byte(`{"server":"ia601607.us.archive.org","dir":"/6/items/nasa"}`), &real)
	if !real.Exists() {
		t.Error("real metadata should Exist")
	}
}

func TestNodeURL(t *testing.T) {
	m := Metadata{Identifier: "nasa", Server: "ia601607.us.archive.org", Dir: "/6/items/nasa"}
	if got := m.NodeURLFor("nasa_meta.xml"); got != "https://ia601607.us.archive.org/6/items/nasa/nasa_meta.xml" {
		t.Errorf("NodeURLFor = %q", got)
	}
	bare := Metadata{Identifier: "nasa"}
	if got := bare.NodeURLFor("x"); got != "https://archive.org/download/nasa/x" {
		t.Errorf("fallback NodeURLFor = %q", got)
	}
}

func TestReplayURL(t *testing.T) {
	if got := ReplayURL("2010", "http://example.com", false); got != "https://web.archive.org/web/2010/http://example.com" {
		t.Errorf("replay = %q", got)
	}
	if got := ReplayURL("2010", "http://example.com", true); got != "https://web.archive.org/web/2010id_/http://example.com" {
		t.Errorf("raw replay = %q", got)
	}
}

func TestAuthHeaderAndMask(t *testing.T) {
	c := &Credentials{Access: "abc", Secret: "0123456789"}
	if got := c.AuthHeader(); got != "LOW abc:0123456789" {
		t.Errorf("auth header = %q", got)
	}
	if got := c.MaskedSecret(); got != "******6789" {
		t.Errorf("masked = %q", got)
	}
	if (&Credentials{}).Valid() {
		t.Error("empty creds should be invalid")
	}
}

func TestUploadHeaders(t *testing.T) {
	h := UploadHeaders(UploadOptions{MakeBucket: true, NoDerive: true, Metadata: map[string]string{"Title": "Hi"}}, 100)
	if h["x-archive-auto-make-bucket"] != "1" {
		t.Error("missing make-bucket header")
	}
	if h["x-archive-queue-derive"] != "0" {
		t.Error("missing no-derive header")
	}
	if h["x-archive-meta-title"] != "Hi" {
		t.Errorf("meta header = %q", h["x-archive-meta-title"])
	}
	if h["x-archive-size-hint"] != "100" {
		t.Errorf("size hint = %q", h["x-archive-size-hint"])
	}
}

func TestValidIdentifier(t *testing.T) {
	for _, ok := range []string{"nasa", "principleofrelat00eins", "CSPAN3_2020", "a.b-c_d"} {
		if !ValidIdentifier(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "has space", "no/slash", "-leadingdash"} {
		if ValidIdentifier(bad) {
			t.Errorf("%q should be invalid", bad)
		}
	}
}

func TestSearchDocStringAndTemplateValue(t *testing.T) {
	var d SearchDoc
	if err := json.Unmarshal([]byte(`{"identifier":"nasa","downloads":90,"collection":["nasa"],"subject":["a","b"]}`), &d); err != nil {
		t.Fatal(err)
	}
	// String coerces scalars, numbers, and first-of-array.
	if got := d.String("identifier"); got != "nasa" {
		t.Errorf("identifier = %q", got)
	}
	if got := d.String("downloads"); got != "90" {
		t.Errorf("downloads = %q", got)
	}
	if got := d.String("subject"); got != "a" {
		t.Errorf("subject = %q", got)
	}

	// TemplateValue decodes to real Go types: strings, numbers, and slices, with
	// single-element arrays collapsed to their element.
	tv, ok := d.TemplateValue().(map[string]any)
	if !ok {
		t.Fatalf("TemplateValue type = %T", d.TemplateValue())
	}
	if tv["identifier"] != "nasa" {
		t.Errorf("tv identifier = %v", tv["identifier"])
	}
	if tv["downloads"] != float64(90) {
		t.Errorf("tv downloads = %v (%T)", tv["downloads"], tv["downloads"])
	}
	if tv["collection"] != "nasa" {
		t.Errorf("tv collection (single-element array) = %v", tv["collection"])
	}
	if arr, ok := tv["subject"].([]any); !ok || len(arr) != 2 {
		t.Errorf("tv subject = %v", tv["subject"])
	}
}

func TestFileInfoPreservesAllFields(t *testing.T) {
	// A real media file record carries fields beyond the typed ones; none must
	// be dropped.
	raw := `{"name":"a.mp3","source":"derivative","format":"VBR MP3","size":"1699742",
		"mtime":"1778378728","md5":"abc","crc32":"def","sha1":"012",
		"title":"00 - Preface","track":"1","album":"Classics","bitrate":"112",
		"length":"02:01","original":"a_128kb.mp3","creator":"McCabe"}`
	var f FileInfo
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatal(err)
	}
	if f.Size() != 1699742 || f.Format != "VBR MP3" {
		t.Errorf("typed fields wrong: size=%d format=%q", f.Size(), f.Format)
	}
	got := f.Fields()
	for _, k := range []string{"title", "track", "album", "bitrate", "length", "original", "creator", "crc32"} {
		if _, ok := got[k]; !ok {
			t.Errorf("Fields() dropped %q", k)
		}
	}
	if got["length"] != "02:01" {
		t.Errorf("length = %v", got["length"])
	}
}

func TestExtractText(t *testing.T) {
	html := []byte(`<html><head><title>T</title><style>x{}</style></head><body><p>Hello</p><script>1</script><p>World</p></body></html>`)
	got := ExtractText(html)
	if got != "Hello\nWorld" {
		t.Errorf("text = %q", got)
	}
	if ExtractTitle(html) != "T" {
		t.Errorf("title = %q", ExtractTitle(html))
	}
}
