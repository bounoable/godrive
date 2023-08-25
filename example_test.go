package godrive_test

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bounoable/godrive"
	"github.com/bounoable/godrive/gcs"
	"github.com/bounoable/godrive/s3"
	"google.golang.org/api/option"
)

func ExampleNew() {
	// Manually configure disks
	manager := godrive.New()

	s3client := awss3.NewFromConfig(aws.Config{})
	manager.Configure("main", s3.NewDisk(s3client, "REGION", "BUCKET", s3.Public(true)), godrive.Default()) // make it the default disk

	gcsclient, _ := storage.NewClient(context.Background(), option.WithCredentialsFile("/path/to/service_account.json"))
	manager.Configure("videos", gcs.NewDisk(gcsclient, "BUCKET", gcs.Public(true)))

	disk, _ := manager.Disk("videos") // Get disk by name

	disk.Put(context.Background(), "path/on/disk.txt", []byte("Hi.")) // Use the disk

	// or use the manager as a disk (uses the "main" disk)
	manager.Put(context.Background(), "path/on/disk.txt", []byte("Hi."))
}

func ExampleNewAutoWire() {
	aw := godrive.NewAutoWire(
		s3.Register,  // Register Amazon S3
		gcs.Register, // Register Google Cloud Storage
	)

	err := aw.Load("/path/to/config.yml") // Load the autowire configuration
	if err != nil {
		panic(err)
	}

	manager, err := aw.NewManager(context.Background()) // Build the storage manager
	if err != nil {
		panic(err)
	}

	disk, _ := manager.Disk("videos") // Get disk by name (as defined by the YAML configuration)

	disk.Put(context.Background(), "path/on/disk.txt", []byte("Hi.")) // Use the disk

	// or use the default disk (as defined by the YAML configuration)
	manager.Put(context.Background(), "path/on/disk.txt", []byte("Hi."))
}
