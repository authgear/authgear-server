package ldap

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Service struct {
	Config       *config.LDAPConfig
	SecretConfig *config.LDAPServerUserCredentials
}
