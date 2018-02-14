// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit

import (
	"regexp"
	"strings"

	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// PasswordViolationReason is a detailed explanation
// of skyerr.PasswordPolicyViolated
type PasswordViolationReason int

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

func (r PasswordViolationReason) String() string {
	switch r {
	case PasswordTooShort:
		return "PasswordTooShort"
	case PasswordUppercaseRequired:
		return "PasswordUppercaseRequired"
	case PasswordLowercaseRequired:
		return "PasswordLowercaseRequired"
	case PasswordDigitRequired:
		return "PasswordDigitRequired"
	case PasswordSymbolRequired:
		return "PasswordSymbolRequired"
	case PasswordContainingExcludedKeywords:
		return "PasswordContainingExcludedKeywords"
	case PasswordBelowGuessableLevel:
		return "PasswordBelowGuessableLevel"
	case PasswordReused:
		return "PasswordReused"
	case PasswordExpired:
		return "PasswordExpired"
	default:
		panic("unreachable")
	}
}

func MakePasswordError(reason PasswordViolationReason, message string, info map[string]interface{}) skyerr.Error {
	newInfo := make(map[string]interface{})
	newInfo["reason"] = reason.String()
	for key, value := range info {
		newInfo[key] = value
	}
	return skyerr.NewErrorWithInfo(skyerr.PasswordPolicyViolated, message, newInfo)
}

func isUpperRune(r rune) bool {
	// NOTE: Intentionally not use unicode.IsUpper
	// because it take other languages into account.
	return r >= 'A' && r <= 'Z'
}

func isLowerRune(r rune) bool {
	// NOTE: Intentionally not use unicode.IsLower
	// because it take other languages into account.
	return r >= 'a' && r <= 'z'
}

func isDigitRune(r rune) bool {
	// NOTE: Intentionally not use unicode.IsDigit
	// because it take other languages into account.
	return r >= '0' && r <= '9'
}

func isSymbolRune(r rune) bool {
	// We define symbol as non-alphanumeric character
	return !isUpperRune(r) && !isLowerRune(r) && !isDigitRune(r)
}

func checkPasswordLength(password string, minLength int) bool {
	if minLength <= 0 {
		return true
	}
	// There exist many ways to define the length of a string
	// For example:
	// 1. The number of bytes of a given encoding
	// 2. The number of code points
	// 3. The number of extended grapheme cluster
	// Here we use the simpliest one:
	// the number of bytes of the given string in UTF-8 encoding
	return len(password) >= minLength
}

func checkPasswordUppercase(password string) bool {
	for _, r := range password {
		if isUpperRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordLowercase(password string) bool {
	for _, r := range password {
		if isLowerRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordDigit(password string) bool {
	for _, r := range password {
		if isDigitRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordSymbol(password string) bool {
	for _, r := range password {
		if isSymbolRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordExcludedKeywords(password string, keywords []string) bool {
	if len(keywords) <= 0 {
		return true
	}
	words := []string{}
	for _, w := range keywords {
		words = append(words, regexp.QuoteMeta(w))
	}
	re, err := regexp.Compile("(?i)" + strings.Join(words, "|"))
	if err != nil {
		return false
	}
	loc := re.FindStringIndex(password)
	if loc == nil {
		return true
	}
	return false
}

func checkPasswordGuessableLevel(password string, minLevel int, userInputs []string) (int, bool) {
	if minLevel <= 0 {
		return 0, true
	}
	minScore := minLevel - 1
	if minScore > 4 {
		minScore = 4
	}
	result := zxcvbn.PasswordStrength(password, userInputs)
	ok := result.Score >= minScore
	return result.Score + 1, ok
}

func userDataToStringStringMap(m map[string]interface{}) map[string]string {
	output := make(map[string]string)
	for key, value := range m {
		str, ok := value.(string)
		if ok {
			output[key] = str
		}
	}
	return output
}

func filterDictionary(m map[string]string, predicate func(string) bool) []string {
	output := []string{}
	for key, value := range m {
		ok := predicate(key)
		if ok {
			output = append(output, value)
		}
	}
	return output
}

func filterDictionaryByKeys(m map[string]string, keys []string) []string {
	lookupMap := make(map[string]bool)
	for _, key := range keys {
		lookupMap[key] = true
	}
	predicate := func(key string) bool {
		_, ok := lookupMap[key]
		return ok
	}

	return filterDictionary(m, predicate)
}

func filterDictionaryTakeAll(m map[string]string) []string {
	predicate := func(key string) bool {
		return true
	}
	return filterDictionary(m, predicate)
}

type ValidatePasswordPayload struct {
	AuthID        string
	PlainPassword string
	UserData      map[string]interface{}
	Conn          skydb.Conn
}

type PasswordChecker struct {
	PwMinLength            int
	PwUppercaseRequired    bool
	PwLowercaseRequired    bool
	PwDigitRequired        bool
	PwSymbolRequired       bool
	PwMinGuessableLevel    int
	PwExcludedKeywords     []string
	PwExcludedFields       []string
	PwHistorySize          int
	PwHistoryDays          int
	PasswordHistoryEnabled bool
}

func (pc *PasswordChecker) checkPasswordLength(password string) skyerr.Error {
	minLength := pc.PwMinLength
	if minLength > 0 && !checkPasswordLength(password, minLength) {
		return MakePasswordError(
			PasswordTooShort,
			"password too short",
			map[string]interface{}{
				"min_length": minLength,
				"pw_length":  len(password),
			},
		)
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordUppercase(password string) skyerr.Error {
	if pc.PwUppercaseRequired && !checkPasswordUppercase(password) {
		return MakePasswordError(
			PasswordUppercaseRequired,
			"password uppercase required",
			nil,
		)
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordLowercase(password string) skyerr.Error {
	if pc.PwLowercaseRequired && !checkPasswordLowercase(password) {
		return MakePasswordError(
			PasswordLowercaseRequired,
			"password lowercase required",
			nil,
		)
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordDigit(password string) skyerr.Error {
	if pc.PwDigitRequired && !checkPasswordDigit(password) {
		return MakePasswordError(
			PasswordDigitRequired,
			"password digit required",
			nil,
		)
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordSymbol(password string) skyerr.Error {
	if pc.PwSymbolRequired && !checkPasswordSymbol(password) {
		return MakePasswordError(
			PasswordSymbolRequired,
			"password symbol required",
			nil,
		)
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordExcludedKeywords(password string) skyerr.Error {
	keywords := pc.PwExcludedKeywords
	if len(keywords) > 0 && !checkPasswordExcludedKeywords(password, keywords) {
		return MakePasswordError(
			PasswordContainingExcludedKeywords,
			"password containing excluded keywords",
			nil,
		)
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordExcludedFields(password string, userData map[string]interface{}) skyerr.Error {
	fields := pc.PwExcludedFields
	if len(fields) > 0 {
		dict := userDataToStringStringMap(userData)
		keywords := filterDictionaryByKeys(dict, fields)
		if !checkPasswordExcludedKeywords(password, keywords) {
			return MakePasswordError(
				PasswordContainingExcludedKeywords,
				"password containing excluded keywords",
				nil,
			)
		}
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordGuessableLevel(password string, userData map[string]interface{}) skyerr.Error {
	minLevel := pc.PwMinGuessableLevel
	if minLevel > 0 {
		dict := userDataToStringStringMap(userData)
		userInputs := filterDictionaryTakeAll(dict)
		level, ok := checkPasswordGuessableLevel(password, minLevel, userInputs)
		if !ok {
			return MakePasswordError(
				PasswordBelowGuessableLevel,
				"password below guessable level",
				map[string]interface{}{
					"min_level": minLevel,
					"pw_level":  level,
				},
			)
		}
	}
	return nil
}

func (pc *PasswordChecker) checkPasswordHistory(password, authID string, conn skydb.Conn) skyerr.Error {
	makeErr := func() skyerr.Error {
		return MakePasswordError(
			PasswordReused,
			"password reused",
			map[string]interface{}{
				"history_size": pc.PwHistorySize,
				"history_days": pc.PwHistoryDays,
			},
		)
	}

	if pc.shouldCheckPasswordHistory() && authID != "" {
		history, err := conn.GetPasswordHistory(
			authID,
			pc.PwHistorySize,
			pc.PwHistoryDays,
		)
		if err != nil {
			return makeErr()
		}
		for _, ph := range history {
			if IsSamePassword(ph.HashedPassword, password) {
				return makeErr()
			}
		}
	}
	return nil
}

func (pc *PasswordChecker) ValidatePassword(payload ValidatePasswordPayload) skyerr.Error {
	password := payload.PlainPassword
	userData := payload.UserData
	conn := payload.Conn
	authID := payload.AuthID
	if err := pc.checkPasswordLength(password); err != nil {
		return err
	}
	if err := pc.checkPasswordUppercase(password); err != nil {
		return err
	}
	if err := pc.checkPasswordLowercase(password); err != nil {
		return err
	}
	if err := pc.checkPasswordDigit(password); err != nil {
		return err
	}
	if err := pc.checkPasswordSymbol(password); err != nil {
		return err
	}
	if err := pc.checkPasswordExcludedKeywords(password); err != nil {
		return err
	}
	if err := pc.checkPasswordExcludedFields(password, userData); err != nil {
		return err
	}
	if err := pc.checkPasswordGuessableLevel(password, userData); err != nil {
		return err
	}
	return pc.checkPasswordHistory(password, authID, conn)
}

func (pc *PasswordChecker) ShouldSavePasswordHistory() bool {
	return pc.PasswordHistoryEnabled
}

func (pc *PasswordChecker) shouldCheckPasswordHistory() bool {
	return pc.ShouldSavePasswordHistory()
}

func IsSamePassword(hashedPassword []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil
}
