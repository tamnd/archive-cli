package ia

import "encoding/json"

// FileInfo is one file record from an item's manifest. The Archive encodes the
// numeric fields as strings; Size/Mtime expose them coerced to int64. The full
// raw record is kept so media-specific fields (length, bitrate, track, album,
// width, height, rotation, ...) survive into json/jsonl/template output.
type FileInfo struct {
	Name   string `json:"name"`
	Source string `json:"source"` // original | derivative | metadata
	Format string `json:"format"`
	MD5    string `json:"md5"`
	CRC32  string `json:"crc32"`
	SHA1   string `json:"sha1"`
	SizeS  string `json:"size"`
	MtimeS string `json:"mtime"`

	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON fills the typed fields and retains the complete raw record so no
// per-file field is ever dropped.
func (f *FileInfo) UnmarshalJSON(b []byte) error {
	type alias FileInfo // sheds UnmarshalJSON to avoid recursion
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*f = FileInfo(a)
	f.Raw = append(f.Raw[:0], b...)
	return nil
}

// Size returns the file size in bytes.
func (f FileInfo) Size() int64 { return atoi64(f.SizeS) }

// Mtime returns the file modification time as Unix seconds.
func (f FileInfo) Mtime() int64 { return atoi64(f.MtimeS) }

// Fields decodes the complete file record into a map so every field the API
// returned is available to json/jsonl/template output, not just the typed ones.
func (f FileInfo) Fields() map[string]any {
	out := map[string]any{}
	if len(f.Raw) > 0 {
		_ = json.Unmarshal(f.Raw, &out)
	}
	return out
}

// Metadata is the decoded Metadata API document for an item. Every top-level key
// the API returns is captured; the full document is also kept in Raw so the
// metadata command emits it byte-for-byte.
type Metadata struct {
	Identifier         string          `json:"-"`
	Meta               MetaDict        `json:"metadata"`
	Files              []FileInfo      `json:"files"`
	Server             string          `json:"server"`
	D1                 string          `json:"d1"`
	D2                 string          `json:"d2"`
	Dir                string          `json:"dir"`
	FilesCount         int             `json:"files_count"`
	ItemSize           int64           `json:"item_size"`
	Created            int64           `json:"created"`
	ItemLastUpdated    int64           `json:"item_last_updated"`
	Uniq               int64           `json:"uniq"`
	IsDark             bool            `json:"is_dark"`
	WorkableServers    []string        `json:"workable_servers"`
	AlternateLocations json.RawMessage `json:"alternate_locations"`
	Raw                json.RawMessage `json:"-"`
}

// Exists reports whether the metadata describes a real item (the API returns an
// empty object for an unknown identifier).
func (m Metadata) Exists() bool { return m.Server != "" || len(m.Files) > 0 || len(m.Meta) > 0 }

// MetaDict is an item's metadata dictionary. Values may be a string or an array
// of strings; Strings flattens either into a slice.
type MetaDict map[string]json.RawMessage

// Get returns the first value for key as a string.
func (d MetaDict) Get(key string) string {
	v := d.Strings(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// Strings returns all values for key as a slice (handling single or array).
func (d MetaDict) Strings(key string) []string {
	raw, ok := d[key]
	if !ok {
		return nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return []string{s}
	}
	var ss []string
	if json.Unmarshal(raw, &ss) == nil {
		return ss
	}
	// Numbers/bools: render verbatim.
	return []string{strings_(raw)}
}

func strings_(raw json.RawMessage) string {
	s := string(raw)
	return s
}

// SearchDoc is one hit from Advanced Search or the Scraping API. Fields are kept
// as raw JSON so any requested column survives; convenience accessors cover the
// common ones.
type SearchDoc map[string]json.RawMessage

// String returns field key as a single string (first element if it is an array).
func (d SearchDoc) String(key string) string {
	raw, ok := d[key]
	if !ok {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var n json.Number
	if json.Unmarshal(raw, &n) == nil {
		return n.String()
	}
	var ss []string
	if json.Unmarshal(raw, &ss) == nil {
		if len(ss) > 0 {
			return ss[0]
		}
		return ""
	}
	return strings_(raw)
}

// Identifier is the item id of a hit.
func (d SearchDoc) Identifier() string { return d.String("identifier") }

// TemplateValue returns a decoded view of the document so Go templates see real
// strings, numbers, and slices instead of raw JSON bytes. Single-element arrays
// collapse to their element, which is what most metadata fields want.
func (d SearchDoc) TemplateValue() any {
	out := make(map[string]any, len(d))
	for k, raw := range d {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			out[k] = string(raw)
			continue
		}
		if arr, ok := v.([]any); ok && len(arr) == 1 {
			v = arr[0]
		}
		out[k] = v
	}
	return out
}

// CDXRecord is one Wayback capture from the CDX server.
type CDXRecord struct {
	URLKey     string `json:"urlkey"`
	Timestamp  string `json:"timestamp"`
	Original   string `json:"original"`
	MimeType   string `json:"mimetype"`
	StatusCode string `json:"statuscode"`
	Digest     string `json:"digest"`
	Length     string `json:"length"`
}

// Snapshot is the closest capture from the Availability API.
type Snapshot struct {
	Available bool   `json:"available"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

// Views holds the short view statistics for an item.
type Views struct {
	AllTime  int64 `json:"all_time"`
	Last30   int64 `json:"last_30day"`
	Last7    int64 `json:"last_7day"`
	HaveData bool  `json:"have_data"`
}
