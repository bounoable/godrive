package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bounoable/godrive"
)

const (
	// Provider is the provider name for Amazon S3.
	Provider = "s3"
)

// Register registeres Amazon S3 as a provider for the disk autowire.
func Register(cfg *godrive.AutoWireConfig) {
	cfg.RegisterProvider(Provider, godrive.DiskCreatorFunc(NewAutoWire))
}

// NewAutoWire creates a new Amazon S3 disk from an autowire configuration.
func NewAutoWire(ctx context.Context, cfg map[string]interface{}) (godrive.Disk, error) {
	if cfg == nil {
		cfg = make(map[string]interface{})
	}

	region, ok := cfg["region"].(string)
	if !ok || region == "" {
		return nil, InvalidConfigValueError{
			Key:     "region",
			Details: "region must be set",
		}
	}

	bucket, ok := cfg["bucket"].(string)
	if !ok || bucket == "" {
		return nil, InvalidConfigValueError{
			Key:     "bucket",
			Details: "storage bucket must be set",
		}
	}

	accessKeyID, ok := cfg["accessKeyId"].(string)
	if !ok || accessKeyID == "" {
		return nil, InvalidConfigValueError{
			Key:     "bucket",
			Details: "accessKeyId must be set",
		}
	}

	secretAccessKey, ok := cfg["secretAccessKey"].(string)
	if !ok || secretAccessKey == "" {
		return nil, InvalidConfigValueError{
			Key:     "bucket",
			Details: "secretAccessKey must be set",
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

	awscfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return NewDisk(s3.NewFromConfig(awscfg), region, bucket, Public(public)), nil
}

// InvalidConfigValueError means the autowire configuration has an invalid config value.
type InvalidConfigValueError struct {
	Key     string
	Details string
}

func (err InvalidConfigValueError) Error() string {
	return fmt.Sprintf("invalid configuration value for key '%s': %s", err.Key, err.Details)
}
