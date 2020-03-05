package password

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
)

type Provider interface {
	principal.Provider
	ValidateLoginID(loginID loginid.LoginID) error
	ValidateLoginIDs(loginIDs []loginid.LoginID) error
	CheckLoginIDKeyType(loginIDKey string, standardKey metadata.StandardKey) bool
	IsRealmValid(realm string) bool
	IsDefaultAllowedRealms() bool
	MakePrincipal(userID string, password string, loginID loginid.LoginID, realm string) (*Principal, error)
	CreatePrincipalsByLoginID(authInfoID string, password string, loginIDs []loginid.LoginID, realm string) ([]*Principal, error)
	CreatePrincipal(principal *Principal) (err error)
	DeletePrincipal(principal *Principal) (err error)
	GetPrincipalByLoginIDWithRealm(loginIDKey string, loginID string, realm string, principal *Principal) (err error)
	GetPrincipalsByUserID(userID string) ([]*Principal, error)
	GetPrincipalsByLoginID(loginIDKey string, loginID string) ([]*Principal, error)
	UpdatePassword(principal *Principal, password string) error
	MigratePassword(principal *Principal, password string) error
}
