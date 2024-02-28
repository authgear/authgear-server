package declarative

import authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"

type NewPasswordData struct {
	TypedData
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

func NewNewPasswordData(d NewPasswordData) NewPasswordData {
	d.Type = DataTypeNewPasswordData
	return d
}

var _ authflow.Data = OAuthData{}

func (NewPasswordData) Data() {}
