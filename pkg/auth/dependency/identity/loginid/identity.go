package loginid

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Identity struct {
	ID              string
	UserID          string
	LoginIDKey      string
	LoginIDType     config.LoginIDKeyType
	LoginID         string
	OriginalLoginID string
	UniqueKey       string
	Claims          map[string]string
}
