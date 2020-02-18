<h1>godrive - Cloud Storage Library</h1>
<p>
  <a href="https://pkg.go.dev/github.com/bounoable/godrive">
    <img alt="GoDoc" src="https://img.shields.io/badge/godoc-reference-purple">
  </a>
  <img alt="Version" src="https://img.shields.io/badge/version-0.1.0-blue.svg?cacheSeconds=2592000" />
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" />
  </a>
</p>

This library provides a uniform access to multiple storage providers and a central configuration for all storage disks.
It autowires the storage disks from a single YAML configuration file.

## Install

```sh
go get github.com/bounoable/godrive
```

## Usage

[**Read the GoDocs (pkg.go.dev)**](https://pkg.go.dev/github.com/bounoable/godrive)

### Autowire from YAML configuration

1. Create configuration

```yaml
default: main # Default disk to use

disks:
  main: # Specify a disk name
    provider: s3 # The storage provider
    config: # Configuration for the storage provider
      region: us-east-2
      bucket: images
      accessKeyId: some-access-key-id
      secretAccessKey: some-secret-access-key
      public: true
  
  videos:
    provider: gcs
    config:
      serviceAccount: /path/to/service/account.json
      bucket: uploads
      public: true
```

2. Create manager

```go
package main

import (
  "github.com/bounoable/godrive"
  "github.com/bounoable/godrive/s3"
  "github.com/bounoable/godrive/gcs"
)

func main() {
    // Initialize autowire & register providers
    aw := godrive.NewAutoWire()

    s3.Register(aw)
    gcs.Register(aw)

    // Load disk configuration
    err := aw.Load("/path/to/config.yml")
    if err != nil {
      panic(err)
    }

    // Create disk manager
    manager, err := aw.NewManager(context.Background())
    if err != nil {
      panic(err)
    }

    // Get disk by name and use it
    disk, _ := manager.Disk("videos")
    err = disk.Put(context.Background(), "path/on/disk.txt", []byte("Hi."))
    content, err := disk.Get(context.Background(), "path/on/disk.txt")
    err = disk.Delete(context.Background(), "path/on/disk.txt")

    // or use the default disk
    err = manager.Put(context.Background(), "path/on/disk.txt", []byte("Hi."))
    content, err := manager.Get(context.Background(), "path/on/disk.txt")
    err = manager.Delete(context.Background(), "path/on/disk.txt")
}
```

### Use without autowire

```go
package main

import (
  "cloud.google.com/go/storage"
  "github.com/aws/aws-sdk-go-v2/aws"
  awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
  "github.com/bounoable/godrive"
  "github.com/bounoable/godrive/s3"
  "github.com/bounoable/godrive/gcs"
  "google.golang.org/api/option"
)

func main() {
    // Create manager and add disks manually
    manager := godrive.New()

    s3client := awss3.New(aws.Config{})
    manager.Configure("main", s3.NewDisk(s3client, "REGION", "BUCKET", s3.Public(true)))

    gcsclient, err := storage.NewClient(context.Backgroud())
    manager.Configure("videos", gcs.NewDisk(gcsclient, "BUCKET", gcs.Public(true)))

    // Get disk by name and use it
    disk, _ := manager.Disk("videos")
    err = disk.Put(context.Background(), "path/on/disk.txt", []byte("Hi."))
    content, err := disk.Get(context.Background(), "path/on/disk.txt")
    err = disk.Delete(context.Background(), "path/on/disk.txt")

    // or use the default disk
    err = manager.Put(context.Background(), "path/on/disk.txt", []byte("Hi."))
    content, err := manager.Get(context.Background(), "path/on/disk.txt")
    err = manager.Delete(context.Background(), "path/on/disk.txt")
}
```
