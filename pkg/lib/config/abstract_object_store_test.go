package config_test

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestAbstractObjectStoreConfigInitialize(t *testing.T) {
	Convey("AbstractObjectStoreConfig.Initialize", t, func() {
		validationCtx := &validation.Context{}

		Convey("should handle empty type", func() {
			c := &config.AbstractObjectStoreConfig{}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should initialize AWS S3 config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeAWSS3,
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should initialize Alibaba Cloud OSS config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeAlibabaCloudOSS,
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should initialize GCP GCS config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeGCPGCS,
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should initialize Azure Blob Storage config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeAzureBlobStorage,
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should initialize MinIO config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeMinIO,
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for invalid type", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: "invalid_type",
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/TYPE: invalid object store type: invalid_type")
		})
	})
}

func TestAbstractObjectStoreConfigValidate(t *testing.T) {
	Convey("AbstractObjectStoreConfig.Validate", t, func() {
		validationCtx := &validation.Context{}

		Convey("should handle empty type", func() {
			c := &config.AbstractObjectStoreConfig{}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should validate AWS S3 config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeAWSS3,
				AWSS3: config.AWSS3ObjectStoreConfig{
					BucketName:      "test-bucket",
					Region:          "us-east-1",
					AccessKeyID:     "test-key",
					SecretAccessKey: "test-secret",
				},
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should validate Alibaba Cloud OSS config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeAlibabaCloudOSS,
				AlibabaCloudOSS: config.AlibabaCloudOSSObjectStoreConfig{
					BucketName:      "test-bucket",
					Region:          "oss-cn-shanghai",
					AccessKeyID:     "test-key",
					SecretAccessKey: "test-secret",
				},
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should validate GCP GCS config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeGCPGCS,
				GCPGCS: config.GCPGCSObjectStoreConfig{
					BucketName:     "test-bucket",
					ServiceAccount: "test@service.com",
				},
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should validate Azure Blob Storage config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeAzureBlobStorage,
				AzureBlobStorage: config.AzureBlobStorageObjectStoreConfig{
					StorageAccount: "test-account",
					Container:      "test-container",
					AccessKey:      "test-key",
				},
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should validate MinIO config", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: config.ObjectStoreTypeMinIO,
				MinIO: config.MinIOObjectStoreConfig{
					Endpoint:        "localhost:9000",
					BucketName:      "test-bucket",
					AccessKeyID:     "test-key",
					SecretAccessKey: "test-secret",
				},
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for invalid type", func() {
			c := &config.AbstractObjectStoreConfig{
				Type: "invalid_type",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/TYPE: invalid object store type: invalid_type")
		})
	})
}

func TestAWSS3ObjectStoreConfigValidate(t *testing.T) {
	Convey("AWSS3ObjectStoreConfig.Validate", t, func() {
		validationCtx := &validation.Context{}

		Convey("should pass with all required fields", func() {
			c := &config.AWSS3ObjectStoreConfig{
				BucketName:      "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for missing bucket name", func() {
			c := &config.AWSS3ObjectStoreConfig{
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/BUCKET_NAME: bucket name must be set")
		})

		Convey("should emit error for missing region", func() {
			c := &config.AWSS3ObjectStoreConfig{
				BucketName:      "test-bucket",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/REGION: region must be set")
		})

		Convey("should emit error for missing access key id", func() {
			c := &config.AWSS3ObjectStoreConfig{
				BucketName:      "test-bucket",
				Region:          "us-east-1",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/ACCESS_KEY_ID: access key id must be set")
		})

		Convey("should emit error for missing secret access key", func() {
			c := &config.AWSS3ObjectStoreConfig{
				BucketName:  "test-bucket",
				Region:      "us-east-1",
				AccessKeyID: "test-key",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/SECRET_ACCESS_KEY: secret key id must be set")
		})
	})
}

func TestAlibabaCloudOSSObjectStoreConfigValidate(t *testing.T) {
	Convey("AlibabaCloudOSSObjectStoreConfig.Validate", t, func() {
		validationCtx := &validation.Context{}

		Convey("should pass with all required fields", func() {
			c := &config.AlibabaCloudOSSObjectStoreConfig{
				BucketName:      "test-bucket",
				Region:          "oss-cn-shanghai",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for missing bucket name", func() {
			c := &config.AlibabaCloudOSSObjectStoreConfig{
				Region:          "oss-cn-shanghai",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/BUCKET_NAME: bucket name must be set")
		})

		Convey("should emit error for missing region", func() {
			c := &config.AlibabaCloudOSSObjectStoreConfig{
				BucketName:      "test-bucket",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/REGION: region must be set")
		})

		Convey("should emit error for missing access key id", func() {
			c := &config.AlibabaCloudOSSObjectStoreConfig{
				BucketName:      "test-bucket",
				Region:          "oss-cn-shanghai",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/ACCESS_KEY_ID: access key id must be set")
		})

		Convey("should emit error for missing secret access key", func() {
			c := &config.AlibabaCloudOSSObjectStoreConfig{
				BucketName:  "test-bucket",
				Region:      "oss-cn-shanghai",
				AccessKeyID: "test-key",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/SECRET_ACCESS_KEY: secret key id must be set")
		})
	})
}

func TestGCPGCSObjectStoreConfigInitialize(t *testing.T) {
	Convey("GCPGCSObjectStoreConfig.Initialize", t, func() {
		validationCtx := &validation.Context{}

		Convey("should handle empty credentials path", func() {
			c := &config.GCPGCSObjectStoreConfig{}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should read credentials JSON file", func() {
			testJSON, err := os.ReadFile("./testdata/gcs_credentials.json")
			So(err, ShouldBeNil)

			c := &config.GCPGCSObjectStoreConfig{
				CredentialsJSONPath: "./testdata/gcs_credentials.json",
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
			So(string(c.CredentialsJSON), ShouldEqual, string(testJSON))
		})

		Convey("should emit error for non-existent credentials file", func() {
			c := &config.GCPGCSObjectStoreConfig{
				CredentialsJSONPath: "./testdata/no_such_file",
			}
			c.Initialize(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/CREDENTIALS_JSON_PATH: open ./testdata/no_such_file: no such file or directory")
		})
	})
}

func TestGCPGCSObjectStoreConfigValidate(t *testing.T) {
	Convey("GCPGCSObjectStoreConfig.Validate", t, func() {
		validationCtx := &validation.Context{}

		Convey("should pass with all required fields", func() {
			c := &config.GCPGCSObjectStoreConfig{
				BucketName:     "test-bucket",
				ServiceAccount: "test@service.com",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for missing bucket name", func() {
			c := &config.GCPGCSObjectStoreConfig{
				ServiceAccount: "test@service.com",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/BUCKET_NAME: bucket name must be set")
		})

		Convey("should emit error for missing service account", func() {
			c := &config.GCPGCSObjectStoreConfig{
				BucketName: "test-bucket",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/SERVICE_ACCOUNT: service account must be set")
		})
	})
}

func TestAzureBlobStorageObjectStoreConfigValidate(t *testing.T) {
	Convey("AzureBlobStorageObjectStoreConfig.Validate", t, func() {
		validationCtx := &validation.Context{}

		Convey("should pass with all required fields", func() {
			c := &config.AzureBlobStorageObjectStoreConfig{
				StorageAccount: "test-account",
				Container:      "test-container",
				AccessKey:      "test-key",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for missing storage account", func() {
			c := &config.AzureBlobStorageObjectStoreConfig{
				Container: "test-container",
				AccessKey: "test-key",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/STORAGE_ACCOUNT: storage account must be set")
		})

		Convey("should emit error for missing container", func() {
			c := &config.AzureBlobStorageObjectStoreConfig{
				StorageAccount: "test-account",
				AccessKey:      "test-key",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/CONTAINER: container must be set")
		})

		Convey("should emit error for missing access key", func() {
			c := &config.AzureBlobStorageObjectStoreConfig{
				StorageAccount: "test-account",
				Container:      "test-container",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/ACCESS_KEY: access key must be set")
		})
	})
}

func TestMinIOObjectStoreConfigValidate(t *testing.T) {
	Convey("MinIOObjectStoreConfig.Validate", t, func() {
		validationCtx := &validation.Context{}

		Convey("should pass with all required fields", func() {
			c := &config.MinIOObjectStoreConfig{
				Endpoint:        "localhost:9000",
				BucketName:      "test-bucket",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeNil)
		})

		Convey("should emit error for missing endpoint", func() {
			c := &config.MinIOObjectStoreConfig{
				BucketName:      "test-bucket",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/ENDPOINT: endpoint must be set")
		})

		Convey("should emit error for missing bucket name", func() {
			c := &config.MinIOObjectStoreConfig{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/BUCKET_NAME: bucket name must be set")
		})

		Convey("should emit error for missing access key id", func() {
			c := &config.MinIOObjectStoreConfig{
				Endpoint:        "localhost:9000",
				BucketName:      "test-bucket",
				SecretAccessKey: "test-secret",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/ACCESS_KEY_ID: access key id must be set")
		})

		Convey("should emit error for missing secret access key", func() {
			c := &config.MinIOObjectStoreConfig{
				Endpoint:    "localhost:9000",
				BucketName:  "test-bucket",
				AccessKeyID: "test-key",
			}
			c.Validate(validationCtx)
			So(validationCtx.Error(""), ShouldBeError, ":\n/SECRET_ACCESS_KEY: secret key id must be set")
		})
	})
}
