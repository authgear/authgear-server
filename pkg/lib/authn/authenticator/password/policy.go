package password

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var PasswordPolicyViolated apierrors.Kind = apierrors.Invalid.WithReason("PasswordPolicyViolated")

type PolicyName string

const (
	// PasswordTooShort is self-explanatory
	PasswordTooShort PolicyName = "PasswordTooShort"
	// PasswordUppercaseRequired means the password does not contain ASCII uppercase character
	PasswordUppercaseRequired PolicyName = "PasswordUppercaseRequired"
	// PasswordLowercaseRequired means the password does not contain ASCII lowercase character
	PasswordLowercaseRequired PolicyName = "PasswordLowercaseRequired"
	// PasswordDigitRequired means the password does not contain ASCII digit character
	PasswordDigitRequired PolicyName = "PasswordDigitRequired"
	// PasswordSymbolRequired means the password does not contain ASCII non-alphanumeric character
	PasswordSymbolRequired PolicyName = "PasswordSymbolRequired"
	// PasswordContainingExcludedKeywords means the password contains configured excluded keywords
	PasswordContainingExcludedKeywords PolicyName = "PasswordContainingExcludedKeywords"
	// PasswordBelowGuessableLevel means the password's guessable level is below configured level.
	// The current implementation uses Dropbox's zxcvbn.
	PasswordBelowGuessableLevel PolicyName = "PasswordBelowGuessableLevel"
	// PasswordReused is self-explanatory
	PasswordReused PolicyName = "PasswordReused"
	// PasswordExpired is self-explanatory
	PasswordExpired PolicyName = "PasswordExpired"
)

type Policy struct {
	Name PolicyName
	Info map[string]interface{} `json:",omitempty"`
}

func (v Policy) Kind() string {
	return string(v.Name)
}
