package audit

import (
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var PasswordPolicyViolated skyerr.Kind = skyerr.Invalid.WithReason("PasswordPolicyViolated")

//go:generate stringer -type=PasswordViolationReason

// PasswordViolationReason is a detailed explanation
// of PasswordPolicyViolated
type PasswordViolationReason int

func (r PasswordViolationReason) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

const (
	// PasswordTooShort is self-explanatory
	PasswordTooShort PasswordViolationReason = iota
	// PasswordUppercaseRequired means the password does not contain ASCII uppercase character
	PasswordUppercaseRequired
	// PasswordLowercaseRequired means the password does not contain ASCII lowercase character
	PasswordLowercaseRequired
	// PasswordDigitRequired means the password does not contain ASCII digit character
	PasswordDigitRequired
	// PasswordSymbolRequired means the password does not contain ASCII non-alphanumeric character
	PasswordSymbolRequired
	// PasswordContainingExcludedKeywords means the password contains configured excluded keywords
	PasswordContainingExcludedKeywords
	// PasswordBelowGuessableLevel means the password's guessable level is below configured level.
	// The current implementation uses Dropbox's zxcvbn.
	PasswordBelowGuessableLevel
	// PasswordReused is self-explanatory
	PasswordReused
	// PasswordExpired is self-explanatory
	PasswordExpired
)

type PasswordViolation struct {
	Reason PasswordViolationReason
	Info   map[string]interface{}
}

func (v PasswordViolation) MarshalJSON() ([]byte, error) {
	d := map[string]interface{}{"kind": v.Reason}
	for k, v := range v.Info {
		d[k] = v
	}
	return json.Marshal(d)
}

type passwordViolations []PasswordViolation

func (passwordViolations) IsTagged(tag errors.DetailTag) bool { return tag == skyerr.APIErrorDetail }
