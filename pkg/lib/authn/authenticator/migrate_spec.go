package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type MigrateSpec struct {
	Type model.AuthenticatorType `json:"type,omitempty"`

	OOBOTP *OOBOTPMigrateSpec `json:"oobotp,omitempty"`
}

func (s *MigrateSpec) GetSpec() *Spec {
	return &Spec{
		Type: s.Type,
		// Support migrate primary authenticator only
		Kind: KindPrimary,
		OOBOTP: &OOBOTPSpec{
			Email: s.OOBOTP.Email,
			Phone: s.OOBOTP.Phone,
		},
	}
}

type OOBOTPMigrateSpec struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}
