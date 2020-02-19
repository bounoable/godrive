package gcs

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
	"github.com/bounoable/godrive"
	"google.golang.org/api/option"
)

const (
	// Provider is the provider name for Google Cloud Storage.
	Provider = "gcs"
)

// Register registeres Google Cloud Storage as a provider for the disk autowire.
func Register(cfg *godrive.AutoWireConfig) {
	cfg.RegisterProvider(Provider, godrive.DiskCreatorFunc(NewAutoWire))
}

// NewAutoWire creates a new Google Cloud Storage disk from an autowire configuration.
func NewAutoWire(ctx context.Context, cfg map[string]interface{}) (godrive.Disk, error) {
	if cfg == nil {
		cfg = make(map[string]interface{})
	}

	serviceAccountPath, ok := cfg["serviceAccount"].(string)
	if !ok {
		return nil, InvalidConfigValueError{
			Key:     "serviceAccount",
			Details: "service account path must be set",
		}
	}

	if _, err := os.Stat(serviceAccountPath); err != nil {
		return nil, InvalidConfigValueError{
			Key:     "serviceAccount",
			Details: fmt.Sprintf("service account file not found: %v", err),
		}
	}

	bucket, ok := cfg["bucket"].(string)
	if !ok || bucket == "" {
		return nil, InvalidConfigValueError{
			Key:     "bucket",
			Details: "storage bucket must be set",
		}
	}

	rpublic, ok := cfg["public"]
	if ok {
		if _, ok := rpublic.(bool); !ok {
			return nil, InvalidConfigValueError{
				Key:     "public",
				Details: fmt.Sprintf("public option must be a boolean but it is '%T'", rpublic),
			}
		}
	}
	public, _ := rpublic.(bool)

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, err
	}

	return NewDisk(client, bucket, Public(public)), nil
}

// InvalidConfigValueError means the autowire configuration has an invalid config value.
type InvalidConfigValueError struct {
	Key     string
	Details string
}

func (err InvalidConfigValueError) Error() string {
	return fmt.Sprintf("invalid configuration value for key '%s': %s", err.Key, err.Details)
}
