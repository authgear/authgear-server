package declarative

import authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"

type NewPasswordData struct {
	TypedData
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`

	// Target which receives the reset password code, only exist on reset password
	MaskedTarget string `json:"masked_target,omitempty"`
}

func NewNewPasswordData(d NewPasswordData) NewPasswordData {
	d.Type = DataTypeNewPasswordData
	return d
}

var _ authflow.Data = OAuthData{}

func (NewPasswordData) Data() {}
