// Package ia is the library behind the archive command line: everything it
// knows about archive.org and the Wayback Machine lives here, with no
// dependency on the command framework.
package ia

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Internet Archive endpoints.
const (
	BaseURL        = "https://archive.org/"
	MetadataURL    = "https://archive.org/metadata/"
	DownloadURL    = "https://archive.org/download/"
	DetailsURL     = "https://archive.org/details/"
	AdvancedSearch = "https://archive.org/advancedsearch.php"
	ScrapeURL      = "https://archive.org/services/search/v1/scrape"
	ViewsURL       = "https://be-api.us.archive.org/views/v1/short/"
	TasksURL       = "https://archive.org/services/tasks.php"
	XAuthnURL      = "https://archive.org/services/xauthn/"
	S3URL          = "https://s3.us.archive.org/"

	// Wayback Machine.
	WaybackAvailURL = "https://archive.org/wayback/available"
	WaybackCDXURL   = "http://web.archive.org/cdx/search/cdx"
	WaybackReplay   = "https://web.archive.org/web/"
	WaybackSaveURL  = "https://web.archive.org/save/"

	// UserAgent identifies the client politely to the Archive's edge.
	UserAgent = "archive-cli/1.0 (+https://github.com/tamnd/archive-cli)"
)

// Defaults for the client and downloader.
const (
	DefaultTimeout = 120 * time.Second
	DefaultRetries = 5
	DefaultDelay   = 250 * time.Millisecond
)

// Config controls library behaviour. The zero value is not usable; call
// DefaultConfig and adjust.
type Config struct {
	DataDir   string
	CacheDir  string
	Workers   int
	Timeout   time.Duration
	Delay     time.Duration
	Retries   int
	UserAgent string
}

// DefaultConfig returns a Config rooted at the data directory with polite
// client defaults.
func DefaultConfig() Config {
	return Config{
		DataDir:   dataDir(),
		CacheDir:  cacheDir(),
		Workers:   defaultWorkers(),
		Timeout:   DefaultTimeout,
		Delay:     DefaultDelay,
		Retries:   DefaultRetries,
		UserAgent: UserAgent,
	}
}

func defaultWorkers() int { return min(max(runtime.NumCPU(), 1), 8) }

// dataDir is the root for everything archive writes: downloads and the cache.
// It defaults to ~/data/archive so all state lives under one predictable tree.
// ARCHIVE_DATA_DIR overrides it.
func dataDir() string {
	if d := os.Getenv("ARCHIVE_DATA_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "data", "archive")
}

// cacheDir holds cached metadata, search pages, and views. It sits under the
// data dir by default; ARCHIVE_CACHE_DIR overrides.
func cacheDir() string {
	if d := os.Getenv("ARCHIVE_CACHE_DIR"); d != "" {
		return d
	}
	return filepath.Join(dataDir(), "cache")
}

// ConfigDir returns the directory holding the credentials file.
func ConfigDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "archive")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "archive")
}

// DownloadDir is where downloaded item files land by default.
func (c Config) DownloadDir() string { return filepath.Join(c.DataDir, "download") }
