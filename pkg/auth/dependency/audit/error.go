package audit

import (
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var PasswordPolicyViolated skyerr.Kind = skyerr.Invalid.WithReason("PasswordPolicyViolated")

type PasswordViolationReason string

const (
	// PasswordTooShort is self-explanatory
	PasswordTooShort PasswordViolationReason = "PasswordTooShort"
	// PasswordUppercaseRequired means the password does not contain ASCII uppercase character
	PasswordUppercaseRequired PasswordViolationReason = "PasswordUppercaseRequired"
	// PasswordLowercaseRequired means the password does not contain ASCII lowercase character
	PasswordLowercaseRequired PasswordViolationReason = "PasswordLowercaseRequired"
	// PasswordDigitRequired means the password does not contain ASCII digit character
	PasswordDigitRequired PasswordViolationReason = "PasswordDigitRequired"
	// PasswordSymbolRequired means the password does not contain ASCII non-alphanumeric character
	PasswordSymbolRequired PasswordViolationReason = "PasswordSymbolRequired"
	// PasswordContainingExcludedKeywords means the password contains configured excluded keywords
	PasswordContainingExcludedKeywords PasswordViolationReason = "PasswordContainingExcludedKeywords"
	// PasswordBelowGuessableLevel means the password's guessable level is below configured level.
	// The current implementation uses Dropbox's zxcvbn.
	PasswordBelowGuessableLevel PasswordViolationReason = "PasswordBelowGuessableLevel"
	// PasswordReused is self-explanatory
	PasswordReused PasswordViolationReason = "PasswordReused"
	// PasswordExpired is self-explanatory
	PasswordExpired PasswordViolationReason = "PasswordExpired"
)

type PasswordViolation struct {
	Reason PasswordViolationReason
	Info   map[string]interface{}
}

func (v PasswordViolation) Kind() string {
	return string(v.Reason)
}

func (v PasswordViolation) MarshalJSON() ([]byte, error) {
	d := map[string]interface{}{"kind": v.Reason}
	for k, v := range v.Info {
		d[k] = v
	}
	return json.Marshal(d)
}
