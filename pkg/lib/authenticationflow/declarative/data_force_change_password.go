package declarative

import authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"

type ForceChangePasswordData struct {
	TypedData
	PasswordPolicy    *PasswordPolicy       `json:"password_policy,omitempty"`
	ForceChangeReason *PasswordChangeReason `json:"force_change_reason,omitempty"`
}

func NewForceChangePasswordData(d ForceChangePasswordData) ForceChangePasswordData {
	d.Type = DataTypeNewPasswordData
	return d
}

var _ authflow.Data = OAuthData{}

func (ForceChangePasswordData) Data() {}
