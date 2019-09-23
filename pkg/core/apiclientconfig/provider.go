package apiclientconfig

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

// Provider provides access to APIClientConfiguration.
type Provider interface {
	Get() (*config.APIClientConfiguration, bool)
	AccessKey(apiKey string) model.AccessKey
}
