package loginid

import "github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"

// TODO(loginid): merge loginid package into this package
type LoginID = loginid.LoginID

type Identity struct {
	ID              string
	UserID          string
	LoginIDKey      string
	LoginID         string
	OriginalLoginID string
	UniqueKey       string
	Claims          map[string]string
}
