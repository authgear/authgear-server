package userverify

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
)

type mockLoginIDProvider struct {
	Identities      []loginid.Identity
	LoginIDKeyTypes map[string]metadata.StandardKey
}

func (p *mockLoginIDProvider) List(userID string) ([]*loginid.Identity, error) {
	var is []*loginid.Identity
	for _, i := range p.Identities {
		if i.UserID == userID {
			ii := i
			is = append(is, &ii)
		}
	}
	return is, nil
}

func (p *mockLoginIDProvider) GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error) {
	var is []*loginid.Identity
	for _, i := range p.Identities {
		if i.LoginID == loginID.Value && (loginID.Key == "" || loginID.Key == i.LoginIDKey) {
			ii := i
			is = append(is, &ii)
		}
	}
	return is, nil
}

func (p *mockLoginIDProvider) IsLoginIDKeyType(loginIDKey string, loginIDKeyType metadata.StandardKey) bool {
	return p.LoginIDKeyTypes[loginIDKey] == loginIDKeyType
}
