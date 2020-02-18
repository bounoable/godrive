package godrive_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bounoable/godrive"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cfg := godrive.NewAutoWire()

	err = cfg.Load(filepath.Join(wd, "testdata/autowire.yml"))
	assert.Nil(t, err)

	tests := []struct {
		name     string
		provider string
		config   map[string]interface{}
	}{
		{
			name:     "googlecloud",
			provider: "gcs",
			config: map[string]interface{}{
				"serviceAccount": "/path/to/service/account.json",
				"bucket":         "uploads",
				"public":         true,
			},
		},
		{
			name:     "amazonaws",
			provider: "s3",
			config: map[string]interface{}{
				"region":          "us-east-2",
				"bucket":          "images",
				"accessKeyId":     "some-access-key-id",
				"secretAccessKey": "some-secret-access-key",
				"public":          true,
			},
		},
		{
			name:     "other",
			provider: "other",
			config:   make(map[string]interface{}),
		},
	}

	for _, test := range tests {
		disk, ok := cfg.Disks[test.name]

		assert.True(t, ok)
		assert.Equal(t, test.provider, disk.Provider)
		assert.Equal(t, test.config, disk.Config)
	}

	assert.Equal(t, "s3", cfg.DefaultDiskName)
}
