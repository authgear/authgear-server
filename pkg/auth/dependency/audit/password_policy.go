package audit

import (
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var PasswordPolicyViolated skyerr.Kind = skyerr.Invalid.WithReason("PasswordPolicyViolated")

type PasswordPolicyName string

const (
	// PasswordTooShort is self-explanatory
	PasswordTooShort PasswordPolicyName = "PasswordTooShort"
	// PasswordUppercaseRequired means the password does not contain ASCII uppercase character
	PasswordUppercaseRequired PasswordPolicyName = "PasswordUppercaseRequired"
	// PasswordLowercaseRequired means the password does not contain ASCII lowercase character
	PasswordLowercaseRequired PasswordPolicyName = "PasswordLowercaseRequired"
	// PasswordDigitRequired means the password does not contain ASCII digit character
	PasswordDigitRequired PasswordPolicyName = "PasswordDigitRequired"
	// PasswordSymbolRequired means the password does not contain ASCII non-alphanumeric character
	PasswordSymbolRequired PasswordPolicyName = "PasswordSymbolRequired"
	// PasswordContainingExcludedKeywords means the password contains configured excluded keywords
	PasswordContainingExcludedKeywords PasswordPolicyName = "PasswordContainingExcludedKeywords"
	// PasswordBelowGuessableLevel means the password's guessable level is below configured level.
	// The current implementation uses Dropbox's zxcvbn.
	PasswordBelowGuessableLevel PasswordPolicyName = "PasswordBelowGuessableLevel"
	// PasswordReused is self-explanatory
	PasswordReused PasswordPolicyName = "PasswordReused"
	// PasswordExpired is self-explanatory
	PasswordExpired PasswordPolicyName = "PasswordExpired"
)

type PasswordPolicy struct {
	Name PasswordPolicyName
	Info map[string]interface{}
}

func (v PasswordPolicy) Kind() string {
	return string(v.Name)
}

func (v PasswordPolicy) MarshalJSON() ([]byte, error) {
	d := map[string]interface{}{"kind": v.Name}
	for k, v := range v.Info {
		d[k] = v
	}
	return json.Marshal(d)
}
