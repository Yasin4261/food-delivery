package domain

import (
	"context"
	"io"
)

// FileStore is the port for storing uploaded files (photos). Implementations
// live in internal/storage (local disk today; an S3-compatible adapter would
// slot in the same way). The adapter owns naming: callers pass content plus an
// extension and receive the public URL path — client-supplied filenames never
// reach the filesystem, which is what makes path traversal impossible by
// construction.
type FileStore interface {
	// Save persists the content under a freshly generated name with the given
	// extension (".jpg", ".png") and returns the public URL path
	// (e.g. "/uploads/3af1….jpg").
	Save(ctx context.Context, ext string, content io.Reader) (string, error)
}
