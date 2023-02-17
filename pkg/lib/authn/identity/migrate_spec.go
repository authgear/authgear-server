package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type MigrateSpec struct {
	Type model.IdentityType `json:"type"`

	LoginID *LoginIDMigrateSpec `json:"login_id,omitempty"`
}

type LoginIDMigrateSpec struct {
	Key   string               `json:"key"`
	Type  model.LoginIDKeyType `json:"type"`
	Value string               `json:"value"`
}
