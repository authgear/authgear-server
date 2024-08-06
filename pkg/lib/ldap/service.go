package ldap

import (
	"github.com/go-ldap/ldap/v3"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Service struct {
	Config       *config.LDAPConfig
	SecretConfig *config.LDAPServerUserCredentials
}

type matchedConfig struct {
	ServerConfig *config.LDAPServerConfig
	SecretConfig *config.LDAPServerUserCredentialsItem
}

func (s *Service) matchServerConfig() []*matchedConfig {
	matchedConfigs := []*matchedConfig{}

	if s.Config == nil || s.SecretConfig == nil {
		return matchedConfigs
	}

	secretConfigMap := make(map[string]*config.LDAPServerUserCredentialsItem)
	for _, item := range s.SecretConfig.Items {
		secretConfigMap[item.Name] = item
	}

	for _, serverConfig := range s.Config.Servers {
		if secretConfig, ok := secretConfigMap[serverConfig.Name]; ok {
			matchedConfigs = append(matchedConfigs, &matchedConfig{
				ServerConfig: serverConfig,
				SecretConfig: secretConfig,
			})
		}
	}
	return matchedConfigs
}

func (s *Service) Authenticate(username string, password string) (*ldap.Entry, error) {
	matchedConfigs := s.matchServerConfig()
	for _, matchedConfig := range matchedConfigs {
		client := NewClient(matchedConfig.ServerConfig, matchedConfig.SecretConfig)
		entry, err := client.AuthenticateUser(username, password)
		if err == nil {
			return entry, nil
		}
	}

	return nil, api.ErrUserNotFound
}
