package userexport

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/cloudstorage"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type UserExportCloudStorage interface {
	PresignPutObject(ctx context.Context, name string, header http.Header) (*http.Request, error)
	PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error)
}

var DependencySet = wire.NewSet(
	NewCloudStorage,
	NewHTTPClient,
	wire.Struct(new(UserExportService), "*"),
	NewLogger,
)

func NewCloudStorage(objectStoreConfig *config.UserExportObjectStoreConfig, c clock.Clock) UserExportCloudStorage {
	switch objectStoreConfig.Type {
	case config.ObjectStoreTypeAWSS3:
		s, err := cloudstorage.NewS3Storage(
			objectStoreConfig.AWSS3.AccessKeyID,
			objectStoreConfig.AWSS3.SecretAccessKey,
			objectStoreConfig.AWSS3.Region,
			objectStoreConfig.AWSS3.BucketName,
		)
		if err != nil {
			panic(err)
		}
		return s
	case config.ObjectStoreTypeGCPGCS:
		s, err := cloudstorage.NewGCSStorage(
			objectStoreConfig.GCPGCS.CredentialsJSON,
			objectStoreConfig.GCPGCS.ServiceAccount,
			objectStoreConfig.GCPGCS.BucketName,
			c,
		)
		if err != nil {
			panic(err)
		}
		return s
	case config.ObjectStoreTypeAzureBlobStorage:
		return cloudstorage.NewAzureStorage(
			objectStoreConfig.AzureBlobStorage.ServiceURL,
			objectStoreConfig.AzureBlobStorage.StorageAccount,
			objectStoreConfig.AzureBlobStorage.AccessKey,
			objectStoreConfig.AzureBlobStorage.Container,
			c,
		)
	case config.ObjectStoreTypeMinIO:
		s, err := cloudstorage.NewMinIOStorage(
			objectStoreConfig.MinIO.Endpoint,
			objectStoreConfig.MinIO.BucketName,
			objectStoreConfig.MinIO.AccessKeyID,
			objectStoreConfig.MinIO.SecretAccessKey,
		)
		if err != nil {
			panic(err)
		}
		return s
	default:
		return nil
	}
}
