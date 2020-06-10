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

package password

import (
	"testing"
	"time"

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
	phData := map[string][]History{
		authID: []History{
			{
				ID:             "1",
				UserID:         authID,
				HashedPassword: []byte("$2a$10$EazYxG5cUdf99wGXDU1fguNxvCe7xQLEgr/Ay6VS9fkkVjHZtpJfm"), // "chima"
				CreatedAt:      time.Date(2017, 11, 3, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:             "2",
				UserID:         authID,
				HashedPassword: []byte("$2a$10$8Z0zqmCZ3pZUlvLD8lN.B.ecN7MX8uVcZooPUFnCcB8tWR6diVc1a"), // "faseng"
				CreatedAt:      time.Date(2017, 11, 2, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:             "3",
				UserID:         authID,
				HashedPassword: []byte("$2a$10$qzmi8TkYosj66xHvc9EfEulKjGoZswJSyNVEmmbLDxNGP/lMm6UXC"), // "coffee"
				CreatedAt:      time.Date(2017, 11, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	phStore := &mockPasswordHistoryStoreImpl{
		Data:    phData,
		TimeNow: func() time.Time { return time.Date(2017, 11, 4, 0, 0, 0, 0, time.UTC) },
	}

	Convey("validate short password", t, func() {
		password := "1"
		pc := &Checker{
			PwMinLength: 2,
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordTooShort, Info: map[string]interface{}{"min_length": 2, "pw_length": 1}},
				},
			},
		)
	})
	Convey("validate uppercase password", t, func() {
		password := "a"
		pc := &Checker{
			PwUppercaseRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordUppercaseRequired},
				},
			},
		)
	})
	Convey("validate lowercase password", t, func() {
		password := "A"
		pc := &Checker{
			PwLowercaseRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordLowercaseRequired},
				},
			},
		)
	})
	Convey("validate digit password", t, func() {
		password := "-"
		pc := &Checker{
			PwDigitRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordDigitRequired},
				},
			},
		)
	})
	Convey("validate symbol password", t, func() {
		password := "azAZ09"
		pc := &Checker{
			PwSymbolRequired: true,
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordSymbolRequired},
				},
			},
		)
	})
	Convey("validate excluded keywords password", t, func() {
		password := "useradmin1"
		pc := &Checker{
			PwExcludedKeywords: []string{"user"},
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordContainingExcludedKeywords},
				},
			},
		)
	})
	Convey("validate excluded fields password", t, func() {
		password := "adalovelace"
		pc := &Checker{
			PwExcludedFields: []string{"first_name"},
		}
		userData := map[string]interface{}{
			"first_name": "Ada",
			"last_name":  "Lovelace",
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
				UserData:      userData,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordContainingExcludedKeywords},
				},
			},
		)
	})
	Convey("validate guessable password", t, func() {
		password := "abcde123456"
		pc := &Checker{
			PwMinGuessableLevel: 5,
		}
		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordBelowGuessableLevel, Info: map[string]interface{}{"min_level": 5, "pw_level": 1}},
				},
			},
		)
	})

	Convey("validate password history", t, func(c C) {
		historySize := 12
		historyDays := 365

		pc := &Checker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
			PasswordHistoryStore:   phStore,
		}

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "chima",
				AuthID:        authID,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordReused, Info: map[string]interface{}{"history_size": historySize, "history_days": historyDays}},
				},
			},
		)

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "coffee",
				AuthID:        authID,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordReused, Info: map[string]interface{}{"history_size": historySize, "history_days": historyDays}},
				},
			},
		)

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "milktea",
				AuthID:        authID,
			}),
			ShouldBeNil,
		)
	})

	Convey("validate password history by size", t, func(c C) {
		historySize := 2
		historyDays := 0

		pc := &Checker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
			PasswordHistoryStore:   phStore,
		}

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "chima",
				AuthID:        authID,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordReused, Info: map[string]interface{}{"history_size": historySize, "history_days": historyDays}},
				},
			},
		)

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "coffee",
				AuthID:        authID,
			}),
			ShouldBeNil,
		)
	})

	Convey("validate password history by days", t, func(c C) {
		historySize := 0
		historyDays := 2

		pc := &Checker{
			PwHistorySize:          historySize,
			PwHistoryDays:          historyDays,
			PasswordHistoryEnabled: true,
			PasswordHistoryStore:   phStore,
		}

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "chima",
				AuthID:        authID,
			}),
			ShouldEqualAPIError,
			PasswordPolicyViolated,
			map[string]interface{}{
				"causes": []skyerr.Cause{
					Policy{Name: PasswordReused, Info: map[string]interface{}{"history_size": historySize, "history_days": historyDays}},
				},
			},
		)

		So(
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: "coffee",
				AuthID:        authID,
			}),
			ShouldBeNil,
		)
	})

	Convey("validate strong password", t, func() {
		// nolint:gosec
		password := "N!hon-no-tsuk!-wa-seka!-1ban-k!re!desu" // 日本の月は世界一番きれいです
		pc := &Checker{
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
			pc.ValidatePassword(ValidatePayload{
				PlainPassword: password,
				UserData:      userData,
			}),
			ShouldEqual,
			nil,
		)
	})
}

func TestPasswordPolicy(t *testing.T) {
	Convey("PasswordPolicy", t, func() {
		Convey("empty", func() {
			pc := &Checker{}
			So(pc.PasswordPolicy(), ShouldBeEmpty)
			So(pc.PasswordPolicy(), ShouldNotBeNil)
		})
		Convey("length", func() {
			pc := &Checker{
				PwMinLength: 8,
			}
			So(pc.PasswordPolicy(), ShouldResemble, []Policy{
				Policy{
					Name: PasswordTooShort,
					Info: map[string]interface{}{
						"min_length": 8,
					},
				},
			})
		})
		Convey("guessable level", func() {
			pc := &Checker{
				PwMinGuessableLevel: 3,
			}
			So(pc.PasswordPolicy(), ShouldResemble, []Policy{
				Policy{
					Name: PasswordBelowGuessableLevel,
					Info: map[string]interface{}{
						"min_level": 3,
					},
				},
			})
		})
		Convey("history", func() {
			pc := &Checker{
				PasswordHistoryEnabled: true,
				PwHistorySize:          10,
				PwHistoryDays:          90,
			}
			So(pc.PasswordPolicy(), ShouldResemble, []Policy{
				Policy{
					Name: PasswordReused,
					Info: map[string]interface{}{
						"history_size": 10,
						"history_days": 90,
					},
				},
			})
		})
		Convey("only output effective policies", func() {
			pc := &Checker{
				PwUppercaseRequired: true,
				PwDigitRequired:     true,
			}
			So(pc.PasswordPolicy(), ShouldResemble, []Policy{
				Policy{
					Name: PasswordUppercaseRequired,
				},
				Policy{
					Name: PasswordDigitRequired,
				},
			})
		})
	})
}
