package deps

import (
	"github.com/google/wire"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	imageshandler "github.com/authgear/authgear-server/pkg/images/handler"
	"github.com/authgear/authgear-server/pkg/lib/cloudstorage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/imageproxy"
)

func NewDirector(extractKey imageproxy.ExtractKey, objectStoreConfig *imagesconfig.ObjectStoreConfig) imageproxy.Director {
	switch objectStoreConfig.Type {
	case imagesconfig.ObjectStoreTypeAWSS3:
		return imageproxy.AWSS3Director{
			ExtractKey: extractKey,
			BucketName: objectStoreConfig.AWSS3.BucketName,
			Region:     objectStoreConfig.AWSS3.Region,
		}
	case imagesconfig.ObjectStoreTypeGCPGCS:
		return imageproxy.GCPGCSDirector{
			ExtractKey: extractKey,
			BucketName: objectStoreConfig.GCPGCS.BucketName,
		}
	case imagesconfig.ObjectStoreTypeAzureBlobStorage:
		return imageproxy.AzureBlobStorageDirector{
			ExtractKey:     extractKey,
			StorageAccount: objectStoreConfig.AzureBlobStorage.StorageAccount,
			Container:      objectStoreConfig.AzureBlobStorage.Container,
		}
	default:
		return nil
	}
}

func NewCloudStorage(objectStoreConfig *imagesconfig.ObjectStoreConfig, c clock.Clock) cloudstorage.Storage {
	switch objectStoreConfig.Type {
	case imagesconfig.ObjectStoreTypeAWSS3:
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
	case imagesconfig.ObjectStoreTypeGCPGCS:
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
	case imagesconfig.ObjectStoreTypeAzureBlobStorage:
		// FIXME(images): azure blob storeage implementation
		return nil
	default:
		return nil
	}
}

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ObjectStoreConfig",
		"LoggerFactory",
		"SentryHub",
		"VipsDaemon",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
	),
	wire.FieldsOf(new(*imagesconfig.EnvironmentConfig),
		"TrustProxy",
	),
	wire.Value(imageshandler.ExtractKey),
	NewDirector,
	clock.DependencySet,
	cloudstorage.DependencySet,
	NewCloudStorage,
)
