// Package gcs provides the Google Cloud Storage disk implementation.
package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	gcs "cloud.google.com/go/storage"
)

// Disk is the Google Cloud Storage disk.
type Disk struct {
	Client *gcs.Client
	Config Config
}

// Config is the disk configuration.
type Config struct {
	Bucket string
	Public bool
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

// NewDisk creates a new Google Cloud Storage disk.
func NewDisk(client *gcs.Client, bucket string, options ...Option) *Disk {
	if client == nil {
		panic("invalid google cloud storage client")
	}

	cfg := Config{Bucket: bucket}
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
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", d.Config.Bucket, path), nil
}
