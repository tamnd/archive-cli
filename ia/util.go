package ia

import (
	"path"
	"regexp"
	"strconv"
	"strings"
)

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

var identifierRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// ValidIdentifier reports whether id is a syntactically valid item identifier.
func ValidIdentifier(id string) bool { return identifierRe.MatchString(id) }

// DownloadURLFor returns the canonical download URL for a file in an item.
func DownloadURLFor(identifier, name string) string {
	return DownloadURL + identifier + "/" + name
}

// NodeURLFor returns the direct datanode URL for a file, given the item's
// metadata (server+dir). It falls back to the canonical download URL when the
// node is unknown.
func (m Metadata) NodeURLFor(name string) string {
	if m.Server == "" || m.Dir == "" {
		return DownloadURLFor(m.Identifier, name)
	}
	return "https://" + m.Server + path.Join("/", m.Dir, name)
}

// DetailsURLFor returns the human details page for an item.
func DetailsURLFor(identifier string) string { return DetailsURL + identifier }

// QuoteQuery wraps a Lucene phrase in quotes when it contains spaces and is not
// already a field:value expression.
func QuoteQuery(s string) string {
	if s == "" || strings.ContainsAny(s, ":\"(") || !strings.Contains(s, " ") {
		return s
	}
	return `"` + s + `"`
}
