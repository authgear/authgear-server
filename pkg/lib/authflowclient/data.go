package authflowclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func Cast(data json.RawMessage, ptr interface{}) error {
	return json.Unmarshal([]byte(data), ptr)
}

func asMap(data json.RawMessage) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func CastVerifyClaimOrCreateAuthenticator(rawData json.RawMessage) (*DataVerifyClaim, *DataCreateAuthenticator, error) {
	dataMap, err := asMap(rawData)
	if err != nil {
		return nil, nil, err
	}

	_, okOTPForm := dataMap["otp_form"]
	_, okOptions := dataMap["options"]

	switch {
	case okOTPForm:
		var data DataVerifyClaim
		err = Cast(rawData, &data)
		if err != nil {
			return nil, nil, err
		}
		return &data, nil, err
	case okOptions:
		var data DataCreateAuthenticator
		err = Cast(rawData, &data)
		if err != nil {
			return nil, nil, err
		}
		return nil, &data, nil
	default:
		return nil, nil, fmt.Errorf("unexpected data: %v", string(rawData))
	}
}

func CastDataChannels(rawData json.RawMessage) (*DataChannels, bool) {
	dataMap, err := asMap(rawData)
	if err != nil {
		return nil, false
	}

	_, okChannels := dataMap["channels"]
	if okChannels {
		var data DataChannels
		err = Cast(rawData, &data)
		if err != nil {
			return nil, false
		}
		return &data, true
	}

	return nil, false
}

func CastForBranch(rawData json.RawMessage) (*DataAuthenticate, *DataCreateAuthenticator, *DataChannels, error) {
	dataMap, err := asMap(rawData)
	if err != nil {
		return nil, nil, nil, err
	}

	_, okDeviceTokenEnabled := dataMap["device_token_enabled"]
	_, okChannels := dataMap["channels"]
	_, okOptions := dataMap["options"]

	switch {
	case okDeviceTokenEnabled:
		var data DataAuthenticate
		err = Cast(rawData, &data)
		if err != nil {
			return nil, nil, nil, err
		}
		return &data, nil, nil, nil
	case okChannels:
		var data DataChannels
		err = Cast(rawData, &data)
		if err != nil {
			return nil, nil, nil, err
		}
		return nil, nil, &data, nil
	case okOptions:
		var data DataCreateAuthenticator
		err = Cast(rawData, &data)
		if err != nil {
			return nil, nil, nil, err
		}
		return nil, &data, nil, nil
	default:
		return nil, nil, nil, fmt.Errorf("unexpected data: %v", string(rawData))
	}
}

type DataPromptCreatePasskey struct {
	CreationOptions *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
}

type DataCreateAuthenticatorTOTP struct {
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
}

type DataPasswordPolicyHistory struct {
	Enabled bool `json:"enabled"`
	Size    int  `json:"size,omitempty"`
	Days    int  `json:"days,omitempty"`
}

type DataPasswordPolicy struct {
	MinimumLength      *int                       `json:"minimum_length,omitempty"`
	UppercaseRequired  bool                       `json:"uppercase_required,omitempty"`
	LowercaseRequired  bool                       `json:"lowercase_required,omitempty"`
	AlphabetRequired   bool                       `json:"alphabet_required,omitempty"`
	DigitRequired      bool                       `json:"digit_required,omitempty"`
	SymbolRequired     bool                       `json:"symbol_required,omitempty"`
	MinimumZxcvbnScore *int                       `json:"minimum_zxcvbn_score,omitempty"`
	History            *DataPasswordPolicyHistory `json:"history,omitempty"`
	ExcludedKeywords   []string                   `json:"excluded_keywords,omitempty"`
}

type DataChangePassword struct {
	PasswordPolicy *DataPasswordPolicy `json:"password_policy,omitempty"`
}

type DataResetPassword struct {
	PasswordPolicy *DataPasswordPolicy `json:"password_policy,omitempty"`
}

type DataOAuth struct {
	Alias                 string                       `json:"alias,omitempty"`
	OAuthProviderType     config.OAuthSSOProviderType  `json:"oauth_provider_type,omitempty"`
	OAuthAuthorizationURL string                       `json:"oauth_authorization_url,omitempty"`
	WechatAppType         config.OAuthSSOWeChatAppType `json:"wechat_app_type,omitempty"`
}

type DataChannels struct {
	Channels         []model.AuthenticatorOOBChannel `json:"channels,omitempty"`
	MaskedClaimValue string                          `json:"masked_claim_value,omitempty"`
}

type DataVerifyClaim struct {
	Channel                        model.AuthenticatorOOBChannel `json:"channel,omitempty"`
	OTPForm                        otp.Form                      `json:"otp_form,omitempty"`
	WebsocketURL                   string                        `json:"websocket_url,omitempty"`
	MaskedClaimValue               string                        `json:"masked_claim_value,omitempty"`
	CodeLength                     int                           `json:"code_length,omitempty"`
	CanResendAt                    time.Time                     `json:"can_resend_at,omitempty"`
	CanCheck                       bool                          `json:"can_check"`
	FailedAttemptRateLimitExceeded bool                          `json:"failed_attempt_rate_limit_exceeded"`
}

type DataCreateAuthenticatorOption struct {
	Authentication Authentication                  `json:"authentication"`
	OTPForm        otp.Form                        `json:"otp_form,omitempty"`
	Channels       []model.AuthenticatorOOBChannel `json:"channels,omitempty"`
	PasswordPolicy *DataPasswordPolicy             `json:"password_policy,omitempty"`
}

type DataCreateAuthenticator struct {
	Options []DataCreateAuthenticatorOption `json:"options,omitempty"`
}

type DataViewRecoveryCode struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

type DataIdentifyOption struct {
	Identification Identification                `json:"identification"`
	ProviderType   string                        `json:"provider_type,omitempty"`
	Alias          string                        `json:"alias,omitempty"`
	WechatAppType  string                        `json:"wechat_app_type,omitempty"`
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`
}

type DataIdentify struct {
	Options []DataIdentifyOption `json:"options,omitempty"`
}

type DataAuthenticateOption struct {
	Authentication    Authentication                  `json:"authentication"`
	OTPForm           otp.Form                        `json:"otp_form,omitempty"`
	MaskedDisplayName string                          `json:"masked_display_name,omitempty"`
	Channels          []model.AuthenticatorOOBChannel `json:"channels,omitempty"`
	RequestOptions    *model.WebAuthnRequestOptions   `json:"request_options,omitempty"`
}

type DataAuthenticate struct {
	Options            []DataAuthenticateOption `json:"options,omitempty"`
	DeviceTokenEnabled bool                     `json:"device_token_enabled"`
}

type DataAccountRecoveryIdentificationOption struct {
	Identification AccountRecoveryIdentification `json:"identification"`
}

type DataAccountRecoveryIdentify struct {
	Options []DataAccountRecoveryIdentificationOption `json:"options,omitempty"`
}
