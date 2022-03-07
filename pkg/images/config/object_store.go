package config

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ObjectStoreConfig struct {
	Type             ObjectStoreType                   `envconfig:"TYPE"`
	AWSS3            AWSS3ObjectStoreConfig            `envconfig:"AWS_S3"`
	GCPGCS           GCPGCSObjectStoreConfig           `envconfig:"GCP_GCS"`
	AzureBlobStorage AzureBlobStorageObjectStoreConfig `envconfig:"AZURE_BLOB_STORAGE"`
}

func (c *ObjectStoreConfig) Validate(ctx *validation.Context) {
	switch c.Type {
	case "":
		break
	case ObjectStoreTypeAWSS3:
		c.AWSS3.Validate(ctx.Child("AWS_S3"))
	case ObjectStoreTypeGCPGCS:
		c.GCPGCS.Validate(ctx.Child("GCP_GCS"))
	case ObjectStoreTypeAzureBlobStorage:
		c.AzureBlobStorage.Validate(ctx.Child("AZURE_BLOB_STORAGE"))
	default:
		ctx.Child("TYPE").EmitErrorMessage(fmt.Sprintf("invalid object store type: %v", c.Type))
	}
}

type ObjectStoreType string

const (
	ObjectStoreTypeAWSS3            ObjectStoreType = "AWS_S3"
	ObjectStoreTypeGCPGCS           ObjectStoreType = "GCP_GCS"
	ObjectStoreTypeAzureBlobStorage ObjectStoreType = "AZURE_BLOB_STORAGE"
)

type AWSS3ObjectStoreConfig struct {
	BucketName string `envconfig:"BUCKET_NAME"`
	Region     string `envconfig:"REGION"`
}

func (c *AWSS3ObjectStoreConfig) Validate(ctx *validation.Context) {
	if c.BucketName == "" {
		ctx.Child("BUCKET_NAME").EmitErrorMessage("bucket name must be set")
	}
	if c.Region == "" {
		ctx.Child("REGION").EmitErrorMessage("region must be set")
	}
}

type GCPGCSObjectStoreConfig struct {
	BucketName string `envconfig:"BUCKET_NAME"`
}

func (c *GCPGCSObjectStoreConfig) Validate(ctx *validation.Context) {
	if c.BucketName == "" {
		ctx.Child("BUCKET_NAME").EmitErrorMessage("bucket name must be set")
	}
}

type AzureBlobStorageObjectStoreConfig struct {
	StorageAccount string `envconfig:"STORAGE_ACCOUNT"`
	Container      string `envconfig:"CONTAINER"`
}

func (c *AzureBlobStorageObjectStoreConfig) Validate(ctx *validation.Context) {
	if c.StorageAccount == "" {
		ctx.Child("STORAGE_ACCOUNT").EmitErrorMessage("storage account must be set")
	}
	if c.Container == "" {
		ctx.Child("CONTAINER").EmitErrorMessage("container must be set")
	}
}
