// Package s3 provides the Amazon S3 disk implementation.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

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
	input := &s3.PutObjectInput{
		Bucket: aws.String(d.Config.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(b),
	}

	if d.Config.Public {
		input.ACL = "public-read"
	}

	if err := input.Validate(); err != nil {
		return err
	}

	req := d.Client.PutObjectRequest(input)
	if _, err := req.Send(ctx); err != nil {
		return err
	}

	return nil
}

// Get retrieves the file with the given key.
func (d *Disk) Get(ctx context.Context, key string) ([]byte, error) {
	req := d.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(d.Config.Bucket),
		Key:    aws.String(key),
	})

	res, err := req.Send(ctx)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(res.Body)
}

// Delete deletes the file with the given key.
func (d *Disk) Delete(ctx context.Context, key string) error {
	req := d.Client.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: aws.String(d.Config.Bucket),
		Key:    aws.String(key),
	})

	if _, err := req.Send(ctx); err != nil {
		return err
	}

	return nil
}

// GetURL returns the public URL for the file at path.
func (d *Disk) GetURL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", d.Config.Bucket, key), nil
}
