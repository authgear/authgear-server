package apiclientconfig

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Provider provides access to APIClientConfiguration.
type Provider interface {
	Get() (*config.APIClientConfiguration, bool)
}
