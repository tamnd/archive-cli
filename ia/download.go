package ia

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DownloadResult reports the outcome of fetching one file.
type DownloadResult struct {
	File     FileInfo
	Path     string
	Bytes    int64
	Skipped  bool // already present with a matching md5
	Verified bool // md5 checked and matched
}

// DownloadFile fetches one file of an item into destDir. When verify is set the
// download is checked against the manifest md5; a pre-existing file with a
// matching md5 is skipped. flat omits any sub-directory in the file name.
func DownloadFile(ctx context.Context, h *HTTPClient, m Metadata, f FileInfo, destDir string, verify, flat bool) (DownloadResult, error) {
	name := f.Name
	if flat {
		name = filepath.Base(name)
	}
	dest := filepath.Join(destDir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return DownloadResult{}, err
	}

	if f.MD5 != "" {
		if sum, err := fileMD5(dest); err == nil && sum == f.MD5 {
			return DownloadResult{File: f, Path: dest, Bytes: f.Size(), Skipped: true, Verified: true}, nil
		}
	}

	resp, err := h.GetDownload(ctx, m.NodeURLFor(f.Name))
	if err != nil {
		return DownloadResult{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 404 {
		return DownloadResult{}, &NotFoundError{URL: m.NodeURLFor(f.Name)}
	}
	if resp.StatusCode/100 != 2 {
		return DownloadResult{}, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, f.Name)
	}

	tmp := dest + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return DownloadResult{}, err
	}
	hasher := md5.New()
	n, err := io.Copy(io.MultiWriter(out, hasher), resp.Body)
	cerr := out.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return DownloadResult{}, err
	}
	if cerr != nil {
		_ = os.Remove(tmp)
		return DownloadResult{}, cerr
	}

	res := DownloadResult{File: f, Path: dest, Bytes: n}
	if verify && f.MD5 != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if got != f.MD5 {
			_ = os.Remove(tmp)
			return DownloadResult{}, fmt.Errorf("md5 mismatch for %s: got %s, want %s", f.Name, got, f.MD5)
		}
		res.Verified = true
	}
	if err := os.Rename(tmp, dest); err != nil {
		return DownloadResult{}, err
	}
	return res, nil
}

func fileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
