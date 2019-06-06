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
	"testing"
	"time"

	ph "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPasswordCheckingFuncs(t *testing.T) {
	Convey("check password length", t, func() {
		So(checkPasswordLength("", 0), ShouldEqual, true)
		So(checkPasswordLength("", 1), ShouldEqual, false)
		So(checkPasswordLength("a", 1), ShouldEqual, true)
		So(checkPasswordLength("ab", 1), ShouldEqual, true)
	})
	Convey("check password uppercase", t, func() {
		So(checkPasswordUppercase("A"), ShouldEqual, true)
		So(checkPasswordUppercase("Z"), ShouldEqual, true)
		So(checkPasswordUppercase("a"), ShouldEqual, false)
	})
	Convey("check password lowercase", t, func() {
		So(checkPasswordLowercase("A"), ShouldEqual, false)
		So(checkPasswordLowercase("a"), ShouldEqual, true)
		So(checkPasswordLowercase("z"), ShouldEqual, true)
	})
	Convey("check password digit", t, func() {
		So(checkPasswordDigit("a"), ShouldEqual, false)
		So(checkPasswordDigit("0"), ShouldEqual, true)
		So(checkPasswordDigit("9"), ShouldEqual, true)
	})
	Convey("check password symbol", t, func() {
		So(checkPasswordSymbol("azAZ09"), ShouldEqual, false)
		So(checkPasswordSymbol("~"), ShouldEqual, true)
	})
	Convey("check password excluded keywords", t, func() {
		p := ".+[]{}^$QuoteRegexMetaCorrectly"
		kws := []string{".", "+", "[", "]", "{", "}", "^", "$"}
		So(checkPasswordExcludedKeywords(p, kws), ShouldEqual, false)

		p = "ADminIsEmbedded"
		kws = []string{"admin"}
		So(checkPasswordExcludedKeywords(p, kws), ShouldEqual, false)

		p = "user"
		kws = []string{"admin", "user"}
		So(checkPasswordExcludedKeywords(p, kws), ShouldEqual, false)

		So(checkPasswordExcludedKeywords(p, nil), ShouldEqual, true)

		p = "a_good_password"
		kws = []string{"bad"}
		So(checkPasswordExcludedKeywords(p, kws), ShouldEqual, true)
	})
	Convey("check password guessable level", t, func() {
		p := "nihongo-wo-manabimashou" // 日本語を学びましょう
		_, ok := checkPasswordGuessableLevel(p, 5, nil)
		So(ok, ShouldEqual, true)

		userInputs := []string{"nihongo", "wo", "manabimashou"}

		_, ok = checkPasswordGuessableLevel(p, 5, userInputs)
		So(ok, ShouldEqual, false)
		_, ok = checkPasswordGuessableLevel(p, 4, userInputs)
		So(ok, ShouldEqual, false)
		_, ok = checkPasswordGuessableLevel(p, 3, userInputs)
		So(ok, ShouldEqual, false)
		_, ok = checkPasswordGuessableLevel(p, 2, userInputs)
		So(ok, ShouldEqual, false)
		_, ok = checkPasswordGuessableLevel(p, 1, userInputs)
		So(ok, ShouldEqual, true)
		_, ok = checkPasswordGuessableLevel(p, 0, userInputs)
		So(ok, ShouldEqual, true)
	})
}

func TestUserDataToStringStringMap(t *testing.T) {
	Convey("user record to map[string]string", t, func() {
		userData := map[string]interface{}{
			"s1":  "s1",
			"s2":  "s2",
			"int": 1,
		}
		So(
			userDataToStringStringMap(userData),
			ShouldResemble,
			map[string]string{
				"s1": "s1",
				"s2": "s2",
			},
		)
	})
}

func TestGetDictionary(t *testing.T) {
	Convey("filter dictionary by keys", t, func() {
		So(
			filterDictionaryByKeys(map[string]string{
				"a": "A",
				"b": "B",
			}, []string{"a"}),
			ShouldEqualStringSliceWithoutOrder,
			[]string{"A"},
		)
	})
	Convey("filter dictionary take all", t, func() {
		So(
			filterDictionaryTakeAll(map[string]string{
				"a": "A",
				"b": "B",
			}),
			ShouldEqualStringSliceWithoutOrder,
			[]string{"A", "B"},
		)
	})
}

func TestValidatePassword(t *testing.T) {
	// fixture
	authID := "chima"
	phData := map[string][]ph.PasswordHistory{
		authID: []ph.PasswordHistory{
			ph.PasswordHistory{
				ID:             "1",
				UserID:         authID,
				HashedPassword: []byte("$2a$10$EazYxG5cUdf99wGXDU1fguNxvCe7xQLEgr/Ay6VS9fkkVjHZtpJfm"), // "chima"
				LoggedAt:       time.Date(2017, 11, 3, 0, 0, 0, 0, time.UTC),
			},
			ph.PasswordHistory{
				ID:             "2",
				UserID:         authID,
				HashedPassword: []byte("$2a$10$8Z0zqmCZ3pZUlvLD8lN.B.ecN7MX8uVcZooPUFnCcB8tWR6diVc1a"), // "faseng"
				LoggedAt:       time.Date(2017, 11, 2, 0, 0, 0, 0, time.UTC),
			},
			ph.PasswordHistory{
				ID:             "3",
				UserID:         authID,
				HashedPassword: []byte("$2a$10$qzmi8TkYosj66xHvc9EfEulKjGoZswJSyNVEmmbLDxNGP/lMm6UXC"), // "coffee"
				LoggedAt:       time.Date(2017, 11, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	phStore := ph.NewMockPasswordHistoryStoreWithData(
		phData,
		func() time.Time { return time.Date(2017, 11, 4, 0, 0, 0, 0, time.UTC) },
	)

	Convey("validate short password", t, func() {
		password := "1"
		pc := &PasswordChecker{
			PwMinLength: 2,
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password too short",
			map[string]interface{}{
				"reason":     PasswordTooShort.String(),
				"min_length": 2,
				"pw_length":  1,
			},
		)
	})
	Convey("validate uppercase password", t, func() {
		password := "a"
		pc := &PasswordChecker{
			PwUppercaseRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password uppercase required",
			map[string]interface{}{
				"reason": PasswordUppercaseRequired.String(),
			},
		)
	})
	Convey("validate lowercase password", t, func() {
		password := "A"
		pc := &PasswordChecker{
			PwLowercaseRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password lowercase required",
			map[string]interface{}{
				"reason": PasswordLowercaseRequired.String(),
			},
		)
	})
	Convey("validate digit password", t, func() {
		password := "-"
		pc := &PasswordChecker{
			PwDigitRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password digit required",
			map[string]interface{}{
				"reason": PasswordDigitRequired.String(),
			},
		)
	})
	Convey("validate symbol password", t, func() {
		password := "azAZ09"
		pc := &PasswordChecker{
			PwSymbolRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password symbol required",
			map[string]interface{}{
				"reason": PasswordSymbolRequired.String(),
			},
		)
	})
	Convey("validate excluded keywords password", t, func() {
		password := "useradmin1"
		pc := &PasswordChecker{
			PwExcludedKeywords: []string{"user"},
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password containing excluded keywords",
			map[string]interface{}{
				"reason": PasswordContainingExcludedKeywords.String(),
			},
		)
	})
	Convey("validate excluded fields password", t, func() {
		password := "adalovelace"
		pc := &PasswordChecker{
			PwExcludedFields: []string{"first_name"},
		}
		userData := map[string]interface{}{
			"first_name": "Ada",
			"last_name":  "Lovelace",
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
				UserData:      userData,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password containing excluded keywords",
			map[string]interface{}{
				"reason": PasswordContainingExcludedKeywords.String(),
			},
		)
	})
	Convey("validate guessable password", t, func() {
		password := "abcde123456"
		pc := &PasswordChecker{
			PwMinGuessableLevel: 5,
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password below guessable level",
			map[string]interface{}{
				"reason":    PasswordBelowGuessableLevel.String(),
				"min_level": 5,
				"pw_level":  1,
			},
		)
	})

	Convey("validate password history", t, func(c C) {
		historySize := 12
		historyDays := 365

		pc := &PasswordChecker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
			PasswordHistoryStore:   phStore,
		}

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "chima",
				AuthID:        authID,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password reused",
			map[string]interface{}{
				"reason":       PasswordReused.String(),
				"history_size": historySize,
				"history_days": historyDays,
			},
		)

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "coffee",
				AuthID:        authID,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password reused",
			map[string]interface{}{
				"reason":       PasswordReused.String(),
				"history_size": historySize,
				"history_days": historyDays,
			},
		)

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "milktea",
				AuthID:        authID,
			}),
			ShouldBeNil,
		)
	})

	Convey("validate password history by size", t, func(c C) {
		historySize := 2
		historyDays := 0

		pc := &PasswordChecker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
			PasswordHistoryStore:   phStore,
		}

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "chima",
				AuthID:        authID,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password reused",
			map[string]interface{}{
				"reason":       PasswordReused.String(),
				"history_size": historySize,
				"history_days": historyDays,
			},
		)

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "coffee",
				AuthID:        authID,
			}),
			ShouldBeNil,
		)
	})

	Convey("validate password history by days", t, func(c C) {
		historySize := 0
		historyDays := 2

		pc := &PasswordChecker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
			PasswordHistoryStore:   phStore,
		}

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "chima",
				AuthID:        authID,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordPolicyViolated,
			"password reused",
			map[string]interface{}{
				"reason":       PasswordReused.String(),
				"history_size": historySize,
				"history_days": historyDays,
			},
		)

		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "coffee",
				AuthID:        authID,
			}),
			ShouldBeNil,
		)
	})

	Convey("validate strong password", t, func() {
		// nolint:gosec
		password := "N!hon-no-tsuk!-wa-seka!-1ban-k!re!desu" // 日本の月は世界一番きれいです
		pc := &PasswordChecker{
			PwMinLength:         8,
			PwUppercaseRequired: true,
			PwLowercaseRequired: true,
			PwDigitRequired:     true,
			PwSymbolRequired:    true,
			PwMinGuessableLevel: 5,
			PwExcludedKeywords:  []string{"user", "admin"},
			PwExcludedFields:    []string{"first_name", "last_name"},
		}
		userData := map[string]interface{}{
			"first_name": "Natsume",
			"last_name":  "Souseki",
		}
		So(
			pc.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
				UserData:      userData,
			}),
			ShouldEqual,
			nil,
		)
	})
}
