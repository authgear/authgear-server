package config

import (
	"io/ioutil"
	"os"

	"github.com/skygeario/skygear-server/pkg/core/redis"
)

// Configuration is Asset Gear configuration.
type Configuration struct {
	Standalone                        bool
	StandaloneTenantConfigurationFile string               `envconfig:"STANDALONE_TENANT_CONFIG_FILE" default:"standalone-tenant-config.yaml"`
	ServerHost                        string               `envconfig:"SERVER_HOST" default:"localhost:3002"`
	Redis                             redis.Configuration  `envconfig:"REDIS"`
	UseInsecureCookie                 bool                 `envconfig:"INSECURE_COOKIE"`
	Storage                           StorageConfiguration `envconfig:"STORAGE"`
}

type StorageBackend string

const (
	StorageBackendAzure StorageBackend = "azure"
	StorageBackendGCS   StorageBackend = "gcs"
	StorageBackendS3    StorageBackend = "s3"
)

type StorageConfiguration struct {
	Backend StorageBackend     `envconfig:"BACKEND"`
	Azure   AzureConfiguration `envconfig:"AZURE"`
	GCS     GCSConfiguration   `envconfig:"GCS"`
	S3      S3Configuration    `envconfig:"S3"`
}

type AzureConfiguration struct {
	StorageAccount string `envconfig:"STORAGE_ACCOUNT"`
	Container      string `envconfig:"CONTAINER"`
	AccessKey      string `envconfig:"ACCESS_KEY"`
}

type GCSConfiguration struct {
	ServiceAccount      string `envconfig:"SERVICE_ACCOUNT"`
	Bucket              string `envconfig:"BUCKET"`
	CredentialsJSONPath string `envconfig:"CREDENTIALS_JSON_PATH"`
	CredentialsJSON     []byte `ignored:"true"`
}

type S3Configuration struct {
	Endpoint  string `envconfig:"ENDPOINT"`
	Region    string `envconfig:"REGION"`
	Bucket    string `envconfig:"BUCKET"`
	AccessKey string `envconfig:"ACCESS_KEY"`
	SecretKey string `envconfig:"SECRET_KEY"`
}

func (c *Configuration) Initialize() error {
	if c.Storage.Backend == StorageBackendGCS {
		p := c.Storage.GCS.CredentialsJSONPath

		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		jsonBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		c.Storage.GCS.CredentialsJSON = jsonBytes
	}
	return nil
}
