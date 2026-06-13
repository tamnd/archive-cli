package ia

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// UploadOptions tune an IAS3 upload.
type UploadOptions struct {
	Metadata    map[string]string // x-archive-meta-<k>:<v> set on item creation
	RemoteName  string            // override the name in the item (default: base of local path)
	MakeBucket  bool              // create the item if it does not exist
	NoDerive    bool              // skip derivation after upload
	ContentType string            // override Content-Type
}

// UploadHeaders builds the IAS3 request headers for an upload (exposed so the
// CLI can show them under --dry-run).
func UploadHeaders(opts UploadOptions, size int64) map[string]string {
	h := map[string]string{}
	if opts.MakeBucket {
		h["x-archive-auto-make-bucket"] = "1"
	}
	if opts.NoDerive {
		h["x-archive-queue-derive"] = "0"
	}
	if size >= 0 {
		h["x-archive-size-hint"] = strconv.FormatInt(size, 10)
	}
	if opts.ContentType != "" {
		h["Content-Type"] = opts.ContentType
	}
	for k, v := range opts.Metadata {
		h["x-archive-meta-"+strings.ToLower(k)] = v
	}
	return h
}

// UploadURL returns the IAS3 target for a file in an item.
func UploadURL(identifier, remoteName string) string {
	return S3URL + identifier + "/" + url.PathEscape(remoteName)
}

// Upload puts one local file into an item over IAS3.
func Upload(ctx context.Context, h *HTTPClient, identifier, localPath string, opts UploadOptions) (string, error) {
	if !ValidIdentifier(identifier) {
		return "", fmt.Errorf("invalid identifier %q", identifier)
	}
	remote := opts.RemoteName
	if remote == "" {
		remote = filepath.Base(localPath)
	}
	f, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	u := UploadURL(identifier, remote)
	resp, err := h.Put(ctx, u, f, info.Size(), UploadHeaders(opts, info.Size()))
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return u, nil
}

// DeleteFile removes a file from an item over IAS3.
func DeleteFile(ctx context.Context, h *HTTPClient, identifier, remoteName string) error {
	resp, err := h.Delete(ctx, UploadURL(identifier, remoteName), map[string]string{"x-archive-cascade-delete": "1"})
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
