// Package s3 provides the Amazon S3 disk implementation.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Disk is the Amazon S3 disk.
type Disk struct {
	Client *s3.Client
	Config Config
}

// Config is the disk configuration.
type Config struct {
	Bucket string
	Region string
	Public bool
}

// Option is a disk configuration option.
type Option func(*Config)

// Public configures the disk to make all uploaded files publicly accessible.
func Public(public bool) Option {
	return func(cfg *Config) {
		cfg.Public = public
	}
}

// NewDisk creates a new Amazon S3 disk.
func NewDisk(client *s3.Client, region, bucket string, options ...Option) *Disk {
	cfg := Config{
		Region: region,
		Bucket: bucket,
	}

	for _, opt := range options {
		opt(&cfg)
	}

	return &Disk{
		Client: client,
		Config: cfg,
	}
}

// Put writes b to the file with the given key.
func (d *Disk) Put(ctx context.Context, key string, b []byte) error {
	return d.PutReader(ctx, key, bytes.NewReader(b))
}

// PutReader writes b to the file with the given key.
func (d *Disk) PutReader(ctx context.Context, key string, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(d.Config.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(b),
	}

	if d.Config.Public {
		input.ACL = "public-read"
	}

	_, err = d.Client.PutObject(ctx, input)

	return err
}

// Get retrieves the file with the given key.
func (d *Disk) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := d.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(obj.Body)
}

// Delete deletes the file with the given key.
func (d *Disk) Delete(ctx context.Context, key string) error {
	_, err := d.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(d.Config.Bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetURL returns the public URL for the file at path.
func (d *Disk) GetURL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", d.Config.Bucket, key), nil
}
