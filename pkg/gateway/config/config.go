package config

import (
	"net/url"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/core/logging"
)

// Configuration is gateway startup configuration
type Configuration struct {
	HTTP struct {
		Host string `envconfig:"HOST" default:"localhost:3001"`
	}
	DB struct {
		ConnectionStr string `envconfig:"DATABASE_URL" required:"true"`
	}
	Router RouterConfig
}

// ReadFromEnv reads from environment variable and update the configuration.
func (c *Configuration) ReadFromEnv() error {
	logger := logging.CreateLogger("gateway")
	if err := godotenv.Load(); err != nil {
		logger.WithError(err).Info(
			"Error in loading .env file, continue without .env")
	}
	err := envconfig.Process("", c)
	return err
}

// RouterConfig contain gears url
type RouterConfig struct {
	AuthGearURL string `envconfig:"AUTH_GEAR_URL" required:"true"`

	routerMap map[string]*url.URL `ignored:"true"`
}

// GetRouterMap provide router map from RouterConfig
func (r *RouterConfig) GetRouterMap() map[string]*url.URL {
	if r.routerMap == nil {
		auth, _ := url.Parse(r.AuthGearURL)
		r.routerMap = map[string]*url.URL{
			"auth": auth,
		}
	}
	return r.routerMap
}
