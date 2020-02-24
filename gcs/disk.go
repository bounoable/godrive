// Package gcs provides the Google Cloud Storage disk implementation.
package gcs

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strings"
	"text/template"

	gcs "cloud.google.com/go/storage"
)

const (
	// DefaultURLTemplate is the default URL template to use for (*Disk).GetURL().
	DefaultURLTemplate = "https://storage.googleapis.com/{{ .Bucket }}/{{ .Path }}"
)

// Disk is the Google Cloud Storage disk.
type Disk struct {
	Client *gcs.Client
	Config Config
}

// Config is the disk configuration.
type Config struct {
	Bucket      string
	Public      bool
	URLTemplate string
}

// Option is a disk configuration option.
type Option func(*Config)

// Public configures the disk to make all uploaded files publicly accessible.
// Note that this option will not work and storing files will return an error
// if "uniform bucket-level access" is configured on the bucket, because
// individual ACL for specific objects is not available for those buckets.
func Public(public bool) Option {
	return func(cfg *Config) {
		cfg.Public = public
	}
}

// URLTemplate overrides the default template for public URLs.
func URLTemplate(tpl string) Option {
	return func(cfg *Config) {
		cfg.URLTemplate = tpl
	}
}

// NewDisk creates a new Google Cloud Storage disk.
func NewDisk(client *gcs.Client, bucket string, options ...Option) *Disk {
	if client == nil {
		panic("invalid google cloud storage client")
	}

	cfg := Config{
		Bucket:      bucket,
		URLTemplate: DefaultURLTemplate,
	}

	for _, opt := range options {
		opt(&cfg)
	}

	return &Disk{
		Client: client,
		Config: cfg,
	}
}

// Put writes b to the file at the given path.
func (d *Disk) Put(ctx context.Context, path string, b []byte) error {
	obj := d.Client.Bucket(d.Config.Bucket).Object(path)

	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, bytes.NewReader(b)); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	if d.Config.Public {
		return d.makePublic(ctx, obj)
	}

	return nil
}

func (d *Disk) makePublic(ctx context.Context, obj *gcs.ObjectHandle) error {
	return obj.ACL().Set(ctx, gcs.AllUsers, gcs.RoleReader)
}

// Get retrieves the file at the given path.
func (d *Disk) Get(ctx context.Context, path string) ([]byte, error) {
	obj := d.Client.Bucket(d.Config.Bucket).Object(path)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

// Delete deletes the file at the given path.
func (d *Disk) Delete(ctx context.Context, path string) error {
	return d.Client.Bucket(d.Config.Bucket).Object(path).Delete(ctx)
}

// GetURL returns the public URL for the file at the given path.
func (d *Disk) GetURL(_ context.Context, path string) (string, error) {
	tpl, err := template.New("url").Parse(d.Config.URLTemplate)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, struct {
		Bucket string
		Path   string
	}{
		Bucket: d.Config.Bucket,
		Path:   path,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
