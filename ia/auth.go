package ia

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoCredentials is returned when an authenticated operation runs without
// configured credentials. The CLI maps it to exit code 4.
var ErrNoCredentials = errors.New("no credentials configured; run 'archive configure'")

// Credentials are the IAS3 access/secret key pair used for authenticated
// reads and all writes.
type Credentials struct {
	Access string
	Secret string
}

// Valid reports whether both keys are present.
func (c *Credentials) Valid() bool { return c != nil && c.Access != "" && c.Secret != "" }

// AuthHeader builds the IAS3 authorization header value.
func (c *Credentials) AuthHeader() string {
	return "LOW " + c.Access + ":" + c.Secret
}

// MaskedSecret returns the secret with all but the last four characters hidden.
func (c *Credentials) MaskedSecret() string {
	if c == nil || c.Secret == "" {
		return ""
	}
	if len(c.Secret) <= 4 {
		return strings.Repeat("*", len(c.Secret))
	}
	return strings.Repeat("*", len(c.Secret)-4) + c.Secret[len(c.Secret)-4:]
}

// CredentialsPath is the file archive configure writes.
func CredentialsPath() string { return filepath.Join(ConfigDir(), "credentials") }

// ResolveCredentials finds credentials in priority order: explicit values,
// environment, then the config file. It always returns a non-nil pointer; call
// Valid to test whether it is usable.
func ResolveCredentials(access, secret string) *Credentials {
	if access != "" && secret != "" {
		return &Credentials{Access: access, Secret: secret}
	}
	a := firstEnv("ARCHIVE_ACCESS_KEY", "IA_ACCESS_KEY")
	s := firstEnv("ARCHIVE_SECRET_KEY", "IA_SECRET_KEY")
	if a != "" && s != "" {
		return &Credentials{Access: a, Secret: s}
	}
	if c, err := LoadCredentials(); err == nil && c.Valid() {
		return c
	}
	return &Credentials{}
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// LoadCredentials reads the credentials file (simple "key = value" lines).
func LoadCredentials() (*Credentials, error) {
	f, err := os.Open(CredentialsPath())
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	c := &Credentials{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		switch key {
		case "access", "access_key", "s3_access":
			c.Access = val
		case "secret", "secret_key", "s3_secret":
			c.Secret = val
		}
	}
	return c, sc.Err()
}

// Save writes the credentials to the config file with mode 0600.
func (c *Credentials) Save() error {
	if !c.Valid() {
		return errors.New("refusing to save empty credentials")
	}
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	content := fmt.Sprintf("# archive credentials (IAS3 keys from https://archive.org/account/s3.php)\naccess = %s\nsecret = %s\n", c.Access, c.Secret)
	return os.WriteFile(CredentialsPath(), []byte(content), 0o600)
}

// xauthnResponse is the shape of the login endpoint reply.
type xauthnResponse struct {
	Success bool `json:"success"`
	Values  struct {
		S3 struct {
			Access string `json:"access"`
			Secret string `json:"secret"`
		} `json:"s3"`
		Email      string `json:"email"`
		ScreenName string `json:"screenname"`
	} `json:"values"`
	Error string `json:"error"`
}

// Login exchanges an email and password for IAS3 keys via the xauthn endpoint.
func Login(ctx context.Context, h *HTTPClient, email, password string) (*Credentials, error) {
	form := url.Values{"email": {email}, "password": {password}}
	u := XAuthnURL + "?op=login"
	body, err := h.PostForm(ctx, u, form)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	var r xauthnResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unexpected login response: %w", err)
	}
	if !r.Success || r.Values.S3.Access == "" {
		if r.Error != "" {
			return nil, fmt.Errorf("login failed: %s", r.Error)
		}
		return nil, errors.New("login failed: check your email and password")
	}
	return &Credentials{Access: r.Values.S3.Access, Secret: r.Values.S3.Secret}, nil
}
