package interaction

import "github.com/skygeario/skygear-server/pkg/core/authn"

type AuthenticatorInfo struct {
	ID            string                  `json:"id"`
	Type          authn.AuthenticatorType `json:"type"`
	Secret        string                  `json:"secret"`
	Props         map[string]interface{}  `json:"props"`
	Authenticator interface{}             `json:"-"`
}

func (i *AuthenticatorInfo) ToSpec() AuthenticatorSpec {
	return AuthenticatorSpec{Type: i.Type, Props: i.Props}
}

func (i *AuthenticatorInfo) ToRef() AuthenticatorRef {
	return AuthenticatorRef{ID: i.ID, Type: i.Type}
}

const (
	// AuthenticatorPropCreatedAt is the creation time of the authenticator
	AuthenticatorPropCreatedAt string = "https://auth.skygear.io/claims/authenticators/created_at"

	// AuthenticatorPropTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorPropTOTPDisplayName string = "https://auth.skygear.io/claims/totp/display_name"

	// AuthenticatorPropOOBOTPID is a claim with string value for OOB authenticator ID.
	AuthenticatorPropOOBOTPID string = "https://auth.skygear.io/claims/oob_otp/id"
	// AuthenticatorPropOOBOTPChannelType is a claim with string value for OOB OTP channel type.
	AuthenticatorPropOOBOTPChannelType string = "https://auth.skygear.io/claims/oob_otp/channel_type"
	// AuthenticatorPropOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorPropOOBOTPEmail string = "https://auth.skygear.io/claims/oob_otp/email"
	// AuthenticatorPropOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorPropOOBOTPPhone string = "https://auth.skygear.io/claims/oob_otp/phone"

	// AuthenticatorPropBearerTokenParentID is a claim with string value for bearer token parent authenticator.
	// nolint:gosec
	AuthenticatorPropBearerTokenParentID string = "https://auth.skygear.io/claims/bearer_token/parent_id"

	// AuthenticatorStateOOBOTPID is a claim with string value for OOB authenticator ID of current interaction.
	AuthenticatorStateOOBOTPID string = AuthenticatorPropOOBOTPID
	// AuthenticatorStateOOBOTPCode is a claim with string value for OOB code of current interaction.
	AuthenticatorStateOOBOTPCode string = "https://auth.skygear.io/claims/oob_otp/code"
	// AuthenticatorStateOOBOTPGenerateTime is a claim with string value for OOB code generate time.
	AuthenticatorStateOOBOTPGenerateTime string = "https://auth.skygear.io/claims/oob_otp/generate_time"
	// AuthenticatorStateOOBOTPTriggerTime is a claim with string value for OOB last trigger time of current interaction.
	AuthenticatorStateOOBOTPTriggerTime string = "https://auth.skygear.io/claims/oob_otp/trigger_time"
)
