package authenticator

const (
	// AuthenticatorPropCreatedAt is the creation time of the authenticator
	AuthenticatorPropCreatedAt string = "https://authgear.com/claims/authenticators/created_at"

	// AuthenticatorPropTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorPropTOTPDisplayName string = "https://authgear.com/claims/totp/display_name"

	// AuthenticatorPropOOBOTPID is a claim with string value for OOB authenticator ID.
	AuthenticatorPropOOBOTPID string = "https://authgear.com/claims/oob_otp/id"
	// AuthenticatorPropOOBOTPChannelType is a claim with string value for OOB OTP channel type.
	AuthenticatorPropOOBOTPChannelType string = "https://authgear.com/claims/oob_otp/channel_type"
	// AuthenticatorPropOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorPropOOBOTPEmail string = "https://authgear.com/claims/oob_otp/email"
	// AuthenticatorPropOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorPropOOBOTPPhone string = "https://authgear.com/claims/oob_otp/phone"
	// AuthenticatorPropOOBOTPPhone is a claim with string value for ID of the bound login ID identity.
	AuthenticatorPropOOBOTPIdentityID string = "https://authgear.com/claims/oob_otp/identity_id"

	// AuthenticatorPropBearerTokenParentID is a claim with string value for bearer token parent authenticator.
	// nolint:gosec
	AuthenticatorPropBearerTokenParentID string = "https://authgear.com/claims/bearer_token/parent_id"

	// AuthenticatorStateOOBOTPID is a claim with string value for OOB authenticator ID of current interaction.
	AuthenticatorStateOOBOTPID string = AuthenticatorPropOOBOTPID
	// AuthenticatorStateOOBOTPCode is a claim with string value for OOB code of current interaction.
	AuthenticatorStateOOBOTPCode string = "https://authgear.com/claims/oob_otp/code"
	// AuthenticatorStateOOBOTPGenerateTime is a claim with string value for OOB code generate time.
	AuthenticatorStateOOBOTPGenerateTime string = "https://authgear.com/claims/oob_otp/generate_time"
	// AuthenticatorStateOOBOTPTriggerTime is a claim with string value for OOB last trigger time of current interaction.
	AuthenticatorStateOOBOTPTriggerTime string = "https://authgear.com/claims/oob_otp/trigger_time"
	// AuthenticatorStateOOBOTPChannelType is a claim with string value for OOB OTP channel type.
	AuthenticatorStateOOBOTPChannelType = AuthenticatorPropOOBOTPChannelType
)
