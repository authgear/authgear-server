package model

const (
	// AMRPWD is from https://tools.ietf.org/html/rfc8176#section-2
	AMRPWD string = "pwd"
	// AMROTP is from https://tools.ietf.org/html/rfc8176#section-2
	AMROTP string = "otp"
	// AMRSMS is from https://tools.ietf.org/html/rfc8176#section-2
	AMRSMS string = "sms"
	// AMRMFA is from https://tools.ietf.org/html/rfc8176#section-2
	AMRMFA string = "mfa"
	// AMRXBiometric exists because rfc8176 does not have a general
	// value for any biometric authentication.
	AMRXBiometric string = "x_biometric"
	// AMRXPasskey exists because rfc8176 does not a general
	// value for passkey.
	AMRXPasskey string = "x_passkey"

	// Unique amrs for each authentication option
	AMRXPrimaryPassword      string = "x_primary_password"
	AMRXPrimaryOOBOTPEmail   string = "x_primary_oob_otp_email"
	AMRXPrimaryOOBOTPSMS     string = "x_primary_oob_otp_sms"
	AMRXPrimaryPasskey       string = "x_primary_passkey"
	AMRXSecondaryPassword    string = "x_secondary_password"
	AMRXSecondaryOOBOTPEmail string = "x_secondary_oob_otp_email"
	AMRXSecondaryOOBOTPSMS   string = "x_secondary_oob_otp_sms"
	AMRXSecondaryTOTP        string = "x_secondary_totp"
	AMRXRecoveryCode         string = "x_recovery_code"
	AMRXDeviceToken          string = "x_device_token"
)
