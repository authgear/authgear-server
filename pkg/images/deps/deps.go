package deps

import (
	"github.com/google/wire"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	imageshandler "github.com/authgear/authgear-server/pkg/images/handler"
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

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ObjectStoreConfig",
		"LoggerFactory",
		"SentryHub",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
	),
	wire.FieldsOf(new(*imagesconfig.EnvironmentConfig),
		"TrustProxy",
	),
	wire.Value(imageshandler.ExtractKey),
	NewDirector,
)
