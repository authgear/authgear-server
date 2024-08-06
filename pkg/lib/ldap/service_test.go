package ldap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestService_matchServerConfig(t *testing.T) {
	Convey("Service.matchServerConfig", t, func() {
		s := &Service{
			Config:       nil,
			SecretConfig: nil,
		}
		matchedConfigs := s.matchServerConfig()
		So(matchedConfigs, ShouldResemble, []*matchedConfig{})

		s.Config = &config.LDAPConfig{
			Servers: []*config.LDAPServerConfig{
				{
					Name:                 "server1",
					URL:                  "ldap://localhost:389",
					BaseDN:               "dc=example,dc=com",
					SearchFilterTemplate: "(uid=%s)",
					UserIDAttributeName:  "uid",
				},
				{
					Name:                 "server2",
					URL:                  "ldap://localhost:389",
					BaseDN:               "dc=example,dc=com",
					SearchFilterTemplate: "(uid=%s)",
					UserIDAttributeName:  "uid",
				},
			},
		}
		s.SecretConfig = &config.LDAPServerUserCredentials{
			Items: []*config.LDAPServerUserCredentialsItem{
				{
					Name:     "server2",
					DN:       "uid=admin,dc=example,dc=com",
					Password: "password",
				},
				{
					Name:     "server1",
					DN:       "uid=admin,dc=example,dc=com",
					Password: "password",
				},
			},
		}

		matchedConfigs = s.matchServerConfig()
		So(matchedConfigs, ShouldResemble, []*matchedConfig{
			{
				ServerConfig: &config.LDAPServerConfig{
					Name:                 "server1",
					URL:                  "ldap://localhost:389",
					BaseDN:               "dc=example,dc=com",
					SearchFilterTemplate: "(uid=%s)",
					UserIDAttributeName:  "uid",
				},
				SecretConfig: &config.LDAPServerUserCredentialsItem{
					Name:     "server1",
					DN:       "uid=admin,dc=example,dc=com",
					Password: "password",
				},
			},
			{
				ServerConfig: &config.LDAPServerConfig{
					Name:                 "server2",
					URL:                  "ldap://localhost:389",
					BaseDN:               "dc=example,dc=com",
					SearchFilterTemplate: "(uid=%s)",
					UserIDAttributeName:  "uid",
				},
				SecretConfig: &config.LDAPServerUserCredentialsItem{
					Name:     "server2",
					DN:       "uid=admin,dc=example,dc=com",
					Password: "password",
				},
			},
		})
	})
}
