package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type MigrateSpec struct {
	Type model.AuthenticatorType `json:"type,omitempty"`

	OOBOTP *OOBOTPMigrateSpec `json:"oobotp,omitempty"`
}

type OOBOTPMigrateSpec struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}
