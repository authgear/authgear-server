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
	// AMRFace is from https://tools.ietf.org/html/rfc8176#section-2
	AMRFace string = "face"
	// AMRXBiometric exists because rfc8176 does not have a general
	// value for any biometric authentication.
	AMRXBiometric string = "x_biometric"
	// AMRXPasskey exists because rfc8176 does not a general
	// value for passkey.
	AMRXPasskey string = "x_passkey"
)
