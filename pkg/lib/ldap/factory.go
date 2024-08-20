package ldap

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type ClientFactory struct {
	Config       *config.LDAPConfig
	SecretConfig *config.LDAPServerUserCredentials
}

func (f *ClientFactory) MakeClient(serverConfig *config.LDAPServerConfig) *Client {
	serverSecret, _ := f.SecretConfig.GetItemByServerName(serverConfig.Name)
	client := NewClient(serverConfig, serverSecret)
	return client
}
