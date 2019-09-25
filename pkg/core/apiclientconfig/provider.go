package apiclientconfig

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

// Provider provides access to APIClientConfiguration.
type Provider interface {
	Get() (string, *config.APIClientConfiguration, bool)
	GetAccessKeyByAPIKey(apiKey string) model.AccessKey
	GetAccessKeyByClientID(clientID string) model.AccessKey
}
