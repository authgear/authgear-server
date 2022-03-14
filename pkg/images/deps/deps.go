package deps

import (
	"github.com/google/wire"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/lib/cloudstorage"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

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
		return cloudstorage.NewAzureStorage(
			objectStoreConfig.AzureBlobStorage.ServiceURL,
			objectStoreConfig.AzureBlobStorage.StorageAccount,
			objectStoreConfig.AzureBlobStorage.AccessKey,
			objectStoreConfig.AzureBlobStorage.Container,
			c,
		)
	default:
		return nil
	}
}

var RootDependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ObjectStoreConfig",
		"LoggerFactory",
		"SentryHub",
		"VipsDaemon",
	),
	wire.FieldsOf(new(*imagesconfig.EnvironmentConfig),
		"TrustProxy",
	),
)

var AppRootDependencySet = wire.NewSet(
	RootDependencySet,
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Config",
	),
)

var RequestDependencySet = wire.NewSet(
	AppRootDependencySet,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
	),
	deps.ProvideRequestContext,
	deps.ProvideRemoteIP,
	deps.ProvideUserAgentString,
	deps.ProvideHTTPHost,
	deps.ProvideHTTPProto,
)

var DependencySet = wire.NewSet(
	RequestDependencySet,
	deps.CommonDependencySet,
	clock.DependencySet,
	cloudstorage.DependencySet,
	NewCloudStorage,
)
