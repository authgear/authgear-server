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

	"github.com/golang/mock/gomock"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
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
	Convey("validate short password", t, func() {
		password := "1"
		ua := &PasswordChecker{
			PwMinLength: 2,
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordTooShort,
			"password too short",
			map[string]interface{}{
				"min_length": 2,
				"pw_length":  1,
			},
		)
	})
	Convey("validate uppercase password", t, func() {
		password := "a"
		ua := &PasswordChecker{
			PwUppercaseRequired: true,
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordUppercaseRequired,
			"password uppercase required",
		)
	})
	Convey("validate lowercase password", t, func() {
		password := "A"
		ua := &PasswordChecker{
			PwLowercaseRequired: true,
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordLowercaseRequired,
			"password lowercase required",
		)
	})
	Convey("validate digit password", t, func() {
		password := "-"
		ua := &PasswordChecker{
			PwDigitRequired: true,
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordDigitRequired,
			"password digit required",
		)
	})
	Convey("validate symbol password", t, func() {
		password := "azAZ09"
		ua := &PasswordChecker{
			PwSymbolRequired: true,
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordSymbolRequired,
			"password symbol required",
		)
	})
	Convey("validate excluded keywords password", t, func() {
		password := "useradmin1"
		ua := &PasswordChecker{
			PwExcludedKeywords: []string{"user"},
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordContainingExcludedKeywords,
			"password containing excluded keywords",
		)
	})
	Convey("validate excluded fields password", t, func() {
		password := "adalovelace"
		ua := &PasswordChecker{
			PwExcludedFields: []string{"first_name"},
		}
		userData := map[string]interface{}{
			"first_name": "Ada",
			"last_name":  "Lovelace",
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
				UserData:      userData,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordContainingExcludedKeywords,
			"password containing excluded keywords",
		)
	})
	Convey("validate guessable password", t, func() {
		password := "abcde123456"
		ua := &PasswordChecker{
			PwMinGuessableLevel: 5,
		}
		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordBelowGuessableLevel,
			"password below guessable level",
			map[string]interface{}{
				"min_level": 5,
				"pw_level":  1,
			},
		)
	})
	Convey("validate password history", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		authID := "chima"
		historySize := 12
		historyDays := 365

		conn := mock_skydb.NewMockConn(ctrl)
		conn.EXPECT().
			GetPasswordHistory(authID, historySize, historyDays).
			MinTimes(1).
			Return([]skydb.PasswordHistory{
				skydb.PasswordHistory{
					ID:             "1",
					AuthID:         authID,
					HashedPassword: []byte("$2a$10$EazYxG5cUdf99wGXDU1fguNxvCe7xQLEgr/Ay6VS9fkkVjHZtpJfm"), // "chima"
					LoggedAt:       time.Date(2017, 11, 1, 0, 0, 0, 0, time.UTC),
				},
			}, nil)

		ua := &PasswordChecker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
		}

		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "chima",
				AuthID:        authID,
				Conn:          conn,
			}),
			ShouldEqualSkyError,
			skyerr.PasswordReused,
			"password reused",
			map[string]interface{}{
				"history_size": historySize,
				"history_days": historyDays,
			},
		)

		So(
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: "faseng",
				AuthID:        authID,
				Conn:          conn,
			}),
			ShouldBeNil,
		)
	})
	Convey("validate strong password", t, func() {
		password := "N!hon-no-tsuk!-wa-seka!-1ban-k!re!desu" // 日本の月は世界一番きれいです
		ua := &PasswordChecker{
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
			ua.ValidatePassword(ValidatePasswordPayload{
				PlainPassword: password,
				UserData:      userData,
			}),
			ShouldEqual,
			nil,
		)
	})
}
