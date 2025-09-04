package authenticator

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type OOBOTP struct {
	ID                   string                  `json:"id"`
	UserID               string                  `json:"user_id"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
	Kind                 string                  `json:"kind"`
	IsDefault            bool                    `json:"is_default"`
	OOBAuthenticatorType model.AuthenticatorType `json:"oob_authenticator_type"`
	Phone                string                  `json:"phone,omitempty"`
	Email                string                  `json:"email,omitempty"`
	Metadata             map[string]interface{}  `json:"metadata,omitempty"`
}

func (a *OOBOTP) ToInfo() *Info {
	return &Info{
		ID:        a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      a.OOBAuthenticatorType,
		Kind:      Kind(a.Kind),
		IsDefault: a.IsDefault,

		OOBOTP: a,
	}
}

func (a *OOBOTP) ToTarget() string {
	switch a.OOBAuthenticatorType {
	case model.AuthenticatorTypeOOBSMS:
		return a.Phone
	case model.AuthenticatorTypeOOBEmail:
		return a.Email
	default:
		panic("authenticator: incompatible authenticator type: " + a.OOBAuthenticatorType)
	}
}

func (a *OOBOTP) ToClaimPair() (claimName model.ClaimName, claimValue string) {
	claimName = a.OOBAuthenticatorType.ToClaimName()
	switch a.OOBAuthenticatorType {
	case model.AuthenticatorTypeOOBSMS:
		return claimName, a.Phone
	case model.AuthenticatorTypeOOBEmail:
		return claimName, a.Email
	default:
		panic("authenticator: incompatible authenticator type: " + a.OOBAuthenticatorType)
	}
}

func (a *OOBOTP) PreferredChannel() *string {
	if a.Metadata == nil {
		return nil
	}
	if preferredChannel, ok := a.Metadata["preferred_channel"].(string); ok {
		return &preferredChannel
	}
	return nil
}

func (a *OOBOTP) SetPreferredChannel(preferredChannel *string) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata["preferred_channel"] = preferredChannel
}
