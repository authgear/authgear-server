package interaction

import "github.com/skygeario/skygear-server/pkg/core/authn"

type IdentityInfo struct {
	ID       string                 `json:"id"`
	Type     authn.IdentityType     `json:"type"`
	Claims   map[string]interface{} `json:"claims"`
	Identity interface{}            `json:"-"`
}

func (i *IdentityInfo) ToSpec() IdentitySpec {
	return IdentitySpec{ID: i.ID, Type: i.Type, Claims: i.Claims}
}

const (
	// IdentityClaimOAuthProvider is a claim with a map value like `{ "type": "azureadv2", "tenant": "test" }`.
	IdentityClaimOAuthProvider string = "https://auth.skygear.io/claims/oauth/provider"
	// IdentityClaimOAuthSubjectID is a claim with a string value like `1098765432`.
	IdentityClaimOAuthSubjectID string = "https://auth.skygear.io/claims/oauth/subject_id"
	// IdentityClaimOAuthData is a claim with a map value containing raw OAuth provider profile.
	IdentityClaimOAuthProfile string = "https://auth.skygear.io/claims/oauth/profile"
	// IdentityClaimOAuthData is a claim with a map value containing mapped OIDC claims.
	IdentityClaimOAuthClaims string = "https://auth.skygear.io/claims/oauth/claims"

	// IdentityClaimLoginIDValue is a claim with a string value indicating the key of login ID.
	IdentityClaimLoginIDKey string = "https://auth.skygear.io/claims/login_id/key"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the value of login ID.
	IdentityClaimLoginIDValue string = "https://auth.skygear.io/claims/login_id/value"
	// IdentityClaimLoginIDUniqueKey is a claim with a string value containing the unique normalized login ID.
	IdentityClaimLoginIDUniqueKey string = "https://auth.skygear.io/claims/login_id/unique_key"
)

type AuthenticatorInfo struct {
	ID            string                 `json:"id"`
	Type          AuthenticatorType      `json:"type"`
	Secret        string                 `json:"secret"`
	Props         map[string]interface{} `json:"props"`
	Authenticator interface{}            `json:"-"`
}

func (i *AuthenticatorInfo) ToSpec() AuthenticatorSpec {
	return AuthenticatorSpec{ID: i.ID, Type: i.Type, Props: i.Props}
}

type AuthenticatorType string

const (
	AuthenticatorTypePassword     AuthenticatorType = "password"
	AuthenticatorTypeTOTP         AuthenticatorType = "totp"
	AuthenticatorTypeOOBOTP       AuthenticatorType = "oob_otp"
	AuthenticatorTypeBearerToken  AuthenticatorType = "bearer_token"
	AuthenticatorTypeRecoveryCode AuthenticatorType = "recovery_code"
)

const (
	// AuthenticatorPropTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorPropTOTPDisplayName string = "https://auth.skygear.io/claims/totp/display_name"

	// AuthenticatorPropOOBOTPChannelType is a claim with string value for OOB OTP channel type.
	AuthenticatorPropOOBOTPChannelType string = "https://auth.skygear.io/claims/oob_otp/channel_type"
	// AuthenticatorPropOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorPropOOBOTPEmail string = "https://auth.skygear.io/claims/oob_otp/email"
	// AuthenticatorPropOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorPropOOBOTPPhone string = "https://auth.skygear.io/claims/oob_otp/phone"

	// AuthenticatorPropBearerTokenParentID is a claim with string value for bearer token parent authenticator.
	// nolint:gosec
	AuthenticatorPropBearerTokenParentID string = "https://auth.skygear.io/claims/bearer_token/parent_id"

	// AuthenticatorStateOOBOTPCode is a claim with string value for OOB authenticator ID of current interaction.
	AuthenticatorStateOOBOTPID string = "https://auth.skygear.io/claims/oob_otp/id"
	// AuthenticatorStateOOBOTPCode is a claim with string value for OOB code of current interaction.
	AuthenticatorStateOOBOTPCode string = "https://auth.skygear.io/claims/oob_otp/code"
	// AuthenticatorStateOOBOTPTriggerTime is a claim with string value for OOB last trigger time of current interaction.
	AuthenticatorStateOOBOTPTriggerTime string = "https://auth.skygear.io/claims/oob_otp/trigger_time"
)
