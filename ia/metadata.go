package ia

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"
)

const metadataTTL = time.Hour

// GetMetadata fetches and decodes the Metadata API document for an item.
func GetMetadata(ctx context.Context, h *HTTPClient, cache *Cache, identifier string) (Metadata, error) {
	u := MetadataURL + identifier
	var body []byte
	if cache != nil {
		if b, ok := cache.Get(u, metadataTTL); ok {
			body = b
		}
	}
	if body == nil {
		b, err := h.FetchBytes(ctx, u)
		if err != nil {
			return Metadata{}, err
		}
		body = b
		if cache != nil {
			cache.Put(u, body)
		}
	}
	var m Metadata
	if err := json.Unmarshal(body, &m); err != nil {
		return Metadata{}, err
	}
	m.Identifier = identifier
	m.Raw = body
	if !m.Exists() {
		return m, &NotFoundError{URL: u}
	}
	return m, nil
}

// GetMetadataSub fetches a sub-resource of the Metadata API (e.g. "files",
// "server", "metadata", or "files/<name>") and returns the raw JSON bytes.
func GetMetadataSub(ctx context.Context, h *HTTPClient, identifier, subpath string) ([]byte, error) {
	u := MetadataURL + identifier
	if subpath != "" {
		u += "/" + strings.TrimPrefix(subpath, "/")
	}
	return h.FetchBytes(ctx, u)
}

// FilterFiles returns the files matching an optional case-insensitive name glob
// and/or an exact-or-substring format match. Empty filters match everything.
func (m Metadata) FilterFiles(glob, format string) []FileInfo {
	var out []FileInfo
	format = strings.ToLower(format)
	for _, f := range m.Files {
		if glob != "" {
			if ok, _ := filepath.Match(glob, f.Name); !ok {
				continue
			}
		}
		if format != "" && !strings.Contains(strings.ToLower(f.Format), format) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// Title returns the item's display title, falling back to the identifier.
func (m Metadata) Title() string {
	if t := m.Meta.Get("title"); t != "" {
		return t
	}
	return m.Identifier
}
