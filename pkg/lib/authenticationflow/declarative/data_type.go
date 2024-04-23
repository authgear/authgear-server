package declarative

type DataType string

const (
	DataTypeIdentificationData                   DataType = "identification_data"
	DataTypeAuthenticationData                   DataType = "authentication_data"
	DataTypeOAuthData                            DataType = "oauth_data"
	DataTypeCreateAuthenticatorData              DataType = "create_authenticator_data"
	DataTypeViewRecoveryCodeData                 DataType = "view_recovery_code_data"
	DataTypeSelectOOBOTPChannelsData             DataType = "select_oob_otp_channels_data"
	DataTypeVerifyOOBOTPData                     DataType = "verify_oob_otp_data"
	DataTypeCreatePasskeyData                    DataType = "create_passkey_data"
	DataTypeCreateTOTPData                       DataType = "create_totp_data"
	DataTypeNewPasswordData                      DataType = "new_password_data"
	DataTypeAccountRecoveryIdentificationData    DataType = "account_recovery_identification_data"
	DataTypeAccountRecoverySelectDestinationData DataType = "account_recovery_select_destination_data"
	DataTypeAccountRecoveryVerifyCodeData        DataType = "account_recovery_verify_code_data"
	DataTypeAccountLinkingIdentificationData     DataType = "account_linking_identification_data"
)

type TypedData struct {
	Type DataType `json:"type,omitempty"`
}
