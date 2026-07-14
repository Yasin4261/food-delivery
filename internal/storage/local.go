// Package storage holds the driven adapters for the domain.FileStore port.
// Local stores uploads on disk; an S3-compatible adapter would implement the
// same interface.
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// URLPrefix is the public path uploads are served under (see the router's
// /uploads route and deploy/Caddyfile).
const URLPrefix = "/uploads/"

// validName matches the names this adapter generates — and nothing else. The
// serving handler uses it to refuse any request that isn't one of our files,
// which shuts out path traversal entirely.
var validName = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png)$`)

// ValidName reports whether name is one this adapter could have generated.
func ValidName(name string) bool { return validName.MatchString(name) }

// Local is a disk-backed domain.FileStore.
type Local struct {
	dir string
}

// NewLocal builds a Local store rooted at dir, creating it if needed.
func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &Local{dir: dir}, nil
}

// Dir returns the root directory (for the serving handler).
func (l *Local) Dir() string { return l.dir }

// Save writes the content under a freshly generated random name with the
// given extension and returns its public URL path. The name never derives
// from client input.
func (l *Local) Save(_ context.Context, ext string, content io.Reader) (string, error) {
	if ext != ".jpg" && ext != ".png" {
		return "", fmt.Errorf("unsupported extension %q", ext)
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate file name: %w", err)
	}
	name := hex.EncodeToString(b[:]) + ext

	f, err := os.OpenFile(filepath.Join(l.dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create upload: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, content); err != nil {
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("write upload: %w", err)
	}
	return URLPrefix + name, nil
}
