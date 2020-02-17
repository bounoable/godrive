package godrive

import "context"

// Disk provides the base cloud storage functions.
type Disk interface {
	// Put writes b to the file at the given path.
	Put(ctx context.Context, path string, b []byte) error
	// Get retrieves the file at the given path.
	Get(ctx context.Context, path string) ([]byte, error)
	// Delete deletes the file at the given path.
	Delete(ctx context.Context, path string) error
}

// URLProvider generates public URLs for files.
type URLProvider interface {
	// GetURL returns the public URL for the file at the given path.
	GetURL(ctx context.Context, path string) (string, error)
}
