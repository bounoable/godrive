package godrive_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bounoable/godrive"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "some-access-key-id")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "some-secret-access-key")

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

func TestNewManager(t *testing.T) {
	cfg := godrive.NewAutoWire()

	cfg.RegisterProvider("test", godrive.DiskCreatorFunc(testDiskCreator))
	cfg.Configure("main", "test", map[string]interface{}{
		"key1": "value1",
		"key2": 2,
		"key3": true,
	})

	m, err := cfg.NewManager(context.Background())
	assert.Nil(t, err)

	disk, err := m.Disk("main")
	assert.Nil(t, err)

	tdisk, ok := disk.(testDisk)
	assert.True(t, ok)

	assert.Equal(t, "value1", tdisk.Key1)
	assert.Equal(t, 2, tdisk.Key2)
	assert.True(t, tdisk.Key3)
}

type testDisk struct {
	Key1 string
	Key2 int
	Key3 bool
}

func (d testDisk) Put(_ context.Context, _ string, _ []byte) error { return nil }
func (d testDisk) Get(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (d testDisk) Delete(_ context.Context, _ string) error        { return nil }

func testDiskCreator(_ context.Context, cfg map[string]interface{}) (godrive.Disk, error) {
	return testDisk{
		Key1: cfg["key1"].(string),
		Key2: cfg["key2"].(int),
		Key3: cfg["key3"].(bool),
	}, nil
}
