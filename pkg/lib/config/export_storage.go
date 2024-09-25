package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type UserExportObjectStoreConfig struct {
	Type             ObjectStoreType                   `envconfig:"TYPE"`
	AWSS3            AWSS3ObjectStoreConfig            `envconfig:"AWS_S3"`
	GCPGCS           GCPGCSObjectStoreConfig           `envconfig:"GCP_GCS"`
	AzureBlobStorage AzureBlobStorageObjectStoreConfig `envconfig:"AZURE_BLOB_STORAGE"`
}

func (c *UserExportObjectStoreConfig) Initialize(ctx *validation.Context) {
	switch c.Type {
	case "":
		break
	case ObjectStoreTypeAWSS3:
		c.AWSS3.Initialize(ctx.Child("AWS_S3"))
	case ObjectStoreTypeGCPGCS:
		c.GCPGCS.Initialize(ctx.Child("GCP_GCS"))
	case ObjectStoreTypeAzureBlobStorage:
		c.AzureBlobStorage.Initialize(ctx.Child("AZURE_BLOB_STORAGE"))
	default:
		ctx.Child("TYPE").EmitErrorMessage(fmt.Sprintf("invalid object store type: %v", c.Type))
	}
}

func (c *UserExportObjectStoreConfig) Validate(ctx *validation.Context) {
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
	BucketName      string `envconfig:"BUCKET_NAME"`
	Region          string `envconfig:"REGION"`
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
}

func (c *AWSS3ObjectStoreConfig) Initialize(ctx *validation.Context) {
}

func (c *AWSS3ObjectStoreConfig) Validate(ctx *validation.Context) {
	if c.BucketName == "" {
		ctx.Child("BUCKET_NAME").EmitErrorMessage("bucket name must be set")
	}
	if c.Region == "" {
		ctx.Child("REGION").EmitErrorMessage("region must be set")
	}
	if c.AccessKeyID == "" {
		ctx.Child("ACCESS_KEY_ID").EmitErrorMessage("access key id must be set")
	}
	if c.SecretAccessKey == "" {
		ctx.Child("SECRET_ACCESS_KEY").EmitErrorMessage("secret key id must be set")
	}
}

type GCPGCSObjectStoreConfig struct {
	BucketName          string `envconfig:"BUCKET_NAME"`
	ServiceAccount      string `envconfig:"SERVICE_ACCOUNT"`
	CredentialsJSONPath string `envconfig:"CREDENTIALS_JSON_PATH"`
	CredentialsJSON     []byte `ignored:"true"`
}

func (c *GCPGCSObjectStoreConfig) Initialize(ctx *validation.Context) {
	var err error
	defer func() {
		if err != nil {
			ctx.Child("CREDENTIALS_JSON_PATH").EmitErrorMessage(err.Error())
		}
	}()

	if c.CredentialsJSONPath == "" {
		return
	}

	p := c.CredentialsJSONPath
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	c.CredentialsJSON = jsonBytes
}

func (c *GCPGCSObjectStoreConfig) Validate(ctx *validation.Context) {
	if c.BucketName == "" {
		ctx.Child("BUCKET_NAME").EmitErrorMessage("bucket name must be set")
	}
	if c.ServiceAccount == "" {
		ctx.Child("SERVICE_ACCOUNT").EmitErrorMessage("service account must be set")
	}
	// In DEV-1689, we support Workload Identity, so service account key is no longer required.
	// if c.CredentialsJSONPath == "" {
	// 	ctx.Child("CREDENTIALS_JSON_PATH").EmitErrorMessage("credentials json must be set")
	// }
}

type AzureBlobStorageObjectStoreConfig struct {
	StorageAccount string `envconfig:"STORAGE_ACCOUNT"`
	Container      string `envconfig:"CONTAINER"`
	// ServiceURL is custom Azure blob storage URL. Empty for default URL.
	ServiceURL string `envconfig:"SERVICE_URL"`
	// AccessKey is encoded in standard BASE64.
	AccessKey string `envconfig:"ACCESS_KEY"`
}

func (c *AzureBlobStorageObjectStoreConfig) Initialize(ctx *validation.Context) {
}

func (c *AzureBlobStorageObjectStoreConfig) Validate(ctx *validation.Context) {
	if c.StorageAccount == "" {
		ctx.Child("STORAGE_ACCOUNT").EmitErrorMessage("storage account must be set")
	}
	if c.Container == "" {
		ctx.Child("CONTAINER").EmitErrorMessage("container must be set")
	}
	if c.AccessKey == "" {
		ctx.Child("ACCESS_KEY").EmitErrorMessage("access key must be set")
	}
}
