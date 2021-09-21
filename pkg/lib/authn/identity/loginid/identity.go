package loginid

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Identity struct {
	ID              string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UserID          string
	LoginIDKey      string
	LoginIDType     config.LoginIDKeyType
	LoginID         string
	OriginalLoginID string
	UniqueKey       string
	Claims          map[string]interface{}
}
