package model

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func GetClientConfig(c []config.OAuthClientConfiguration, clientID string) (config.OAuthClientConfiguration, bool) {
	for _, clientConfig := range c {
		if clientConfig.ClientID() == clientID {
			cc := clientConfig
			return cc, true
		}
	}
	return nil, false
}
