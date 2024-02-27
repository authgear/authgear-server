package declarative

type DataType string

const (
	DataTypeIdentificationData                   DataType = "identification_data"
	DataTypeAuthenticationData                   DataType = "authentication_data"
	DataTypeOAuthData                            DataType = "oauth_data"
	DataTypeCreateAuthenticatorData              DataType = "create_authenticator_data"
	DataTypeViewRecoveryCodeData                 DataType = "view_recovery_code_data"
	DataTypeOOBChannelsData                      DataType = "oob_channels_data"
	DataTypeVerifyOOBOTPData                     DataType = "verify_oob_otp_data"
	DataTypeCreatePasskeyData                    DataType = "create_passkey_data"
	DataTypeCreateTOTPData                       DataType = "create_totp_data"
	DataTypeChangePasswordData                   DataType = "change_password_data"
	DataTypeAccountRecoveryIdentificationData    DataType = "account_recovery_identification_data"
	DataTypeAccountRecoverySelectDestinationData DataType = "account_recovery_select_destination_data"
	DataTypeAccountRecoveryVerifyCodeData        DataType = "account_recovery_verify_code_data"
	DataTypeResetPasswordData                    DataType = "reset_password_data"
)

type TypedData struct {
	Type DataType `json:"type,omitempty"`
}
