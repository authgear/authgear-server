package ldap

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type ClientFactory struct {
	Config       *config.LDAPConfig
	SecretConfig *config.LDAPServerUserCredentials
}

func (f *ClientFactory) Authenticate(serverName string, username string, password string) (*Entry, error) {
	var serverConfig *config.LDAPServerConfig
	for _, s := range f.Config.Servers {
		if s.Name == serverName {
			serverConfig = s
			break
		}
	}

	if serverConfig == nil {
		return nil, api.ErrLDAPServerNotFound
	}

	var serverSecret config.LDAPServerUserCredentialsItem
	for _, s := range f.SecretConfig.Items {
		if s.Name == serverName {
			serverSecret = s
		}
	}
	client := NewClient(serverConfig, &serverSecret)
	ldapEntry, err := client.AuthenticateUser(username, password)
	if err != nil {
		return nil, err
	}
	return &Entry{
		ldapEntry,
	}, nil
}
