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

package skydb

import (
	"bytes"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestNewAuthInfo(t *testing.T) {
	info := NewAuthInfo("secret")

	if info.ID == "" {
		t.Fatalf("got empty info.ID, want non-empty string")
	}

	if bytes.Equal(info.HashedPassword, nil) {
		t.Fatalf("got info.HashPassword = %v, want non-empty value", info.HashedPassword)
	}
}

func TestNewAnonymousAuthInfo(t *testing.T) {
	info := NewAnonymousAuthInfo()

	if info.ID == "" {
		t.Fatalf("got info.ID = %v, want \"\"", info.ID)
	}

	if len(info.HashedPassword) != 0 {
		t.Fatalf("got info.HashPassword = %v, want zero-length bytes", info.HashedPassword)
	}
}

func TestNewProviderInfoAuthInfo(t *testing.T) {
	k := "com.example:johndoe"
	v := map[string]interface{}{
		"hello": "world",
	}

	Convey("Test Provied ProviderInfo", t, func() {
		info := NewProviderInfoAuthInfo(k, v)
		So(info.ProviderInfo[k], ShouldResemble, v)
		So(len(info.HashedPassword), ShouldEqual, 0)
	})
}

func TestAuthData(t *testing.T) {
	Convey("Test AuthData", t, func() {
		keys := [][]string{[]string{"username"}, []string{"email"}}

		Convey("valid AuthData", func() {
			So(NewAuthData(map[string]interface{}{
				"username": "johndoe",
			}, keys).IsValid(), ShouldBeTrue)

			So(NewAuthData(map[string]interface{}{
				"email": "johndoe@example.com",
			}, keys).IsValid(), ShouldBeTrue)

			authData := NewAuthData(map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}, keys)
			So(authData.IsValid(), ShouldBeTrue)
			So(authData.usingKeys(), ShouldResemble, []string{"username"})
		})

		Convey("invalid AuthData", func() {
			So(NewAuthData(map[string]interface{}{}, keys).IsValid(), ShouldBeFalse)
			So(NewAuthData(map[string]interface{}{
				"iamyourfather": "johndoe",
			}, keys).IsValid(), ShouldBeFalse)
			So(NewAuthData(map[string]interface{}{
				"username": nil,
			}, keys).IsValid(), ShouldBeFalse)
		})

		Convey("empty AuthData", func() {
			So(NewAuthData(map[string]interface{}{}, keys).IsEmpty(), ShouldBeTrue)
			So(NewAuthData(map[string]interface{}{
				"username": nil,
			}, keys).IsEmpty(), ShouldBeTrue)
			So(NewAuthData(map[string]interface{}{
				"iamyourfather": "johndoe",
			}, keys).IsEmpty(), ShouldBeFalse)
		})
	})

	Convey("Test AuthData with multiple keys", t, func() {
		keys := [][]string{[]string{"username", "email"}, []string{"username", "phone"}}

		Convey("valid AuthData", func() {
			So(NewAuthData(map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}, keys).IsValid(), ShouldBeTrue)

			So(NewAuthData(map[string]interface{}{
				"username": "johndoe",
				"phone":    "123",
			}, keys).IsValid(), ShouldBeTrue)

			authData := NewAuthData(map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
				"phone":    "123",
			}, keys)
			So(authData.IsValid(), ShouldBeTrue)
			So(authData.usingKeys(), ShouldResemble, []string{"username", "email"})
		})

		Convey("invalid AuthData", func() {
			So(NewAuthData(map[string]interface{}{}, keys).IsValid(), ShouldBeFalse)
			So(NewAuthData(map[string]interface{}{
				"username": "johndoe",
			}, keys).IsValid(), ShouldBeFalse)
			So(NewAuthData(map[string]interface{}{
				"email": "johndoe@example.com",
			}, keys).IsValid(), ShouldBeFalse)
			So(NewAuthData(map[string]interface{}{
				"phone": "123",
			}, keys).IsValid(), ShouldBeFalse)
			So(NewAuthData(map[string]interface{}{
				"email": "johndoe@example.com",
				"phone": "123",
			}, keys).IsValid(), ShouldBeFalse)
		})
	})
}

func TestSetPassword(t *testing.T) {
	info := AuthInfo{}
	info.SetPassword("secret")
	err := bcrypt.CompareHashAndPassword(info.HashedPassword, []byte("secret"))
	if err != nil {
		t.Fatalf("got err = %v, want nil", err)
	}
	if info.TokenValidSince == nil {
		t.Fatalf("got info.TokenValidSince = nil, want non-nil")
	}
	if info.TokenValidSince.IsZero() {
		t.Fatalf("got info.TokenValidSince.IsZero = true, want false")
	}
	if !info.IsPasswordSet {
		t.Fatalf("got info.IsPasswordSet = false, want true")
	}
}

func TestIsPasswordExpired(t *testing.T) {
	now := time.Date(2017, 12, 2, 0, 0, 0, 0, time.UTC)
	Convey("IsPasswordExpired", t, func() {
		Convey("return false if authinfo is not password based", func() {
			info := AuthInfo{}
			So(info.IsPasswordExpired(1, now), ShouldBeFalse)
		})
		Convey("return false if expiryDays is not positive", func() {
			info := AuthInfo{}
			info.HashedPassword = []byte("unimportant")
			tokenValidSince := time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC)
			info.TokenValidSince = &tokenValidSince
			So(info.IsPasswordExpired(0, now), ShouldBeFalse)
		})
		Convey("return false if password is indeed valid", func() {
			info := AuthInfo{}
			info.HashedPassword = []byte("unimportant")
			tokenValidSince := time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC)
			info.TokenValidSince = &tokenValidSince
			So(info.IsPasswordExpired(30, now), ShouldBeFalse)
		})
		Convey("return true if password is indeed expired", func() {
			info := AuthInfo{}
			info.HashedPassword = []byte("unimportant")
			tokenValidSince := time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC)
			info.TokenValidSince = &tokenValidSince
			So(info.IsPasswordExpired(30, info.TokenValidSince.AddDate(0, 0, 29)), ShouldBeFalse)
			So(info.IsPasswordExpired(30, info.TokenValidSince.AddDate(0, 0, 30)), ShouldBeFalse)
			So(info.IsPasswordExpired(30, info.TokenValidSince.AddDate(0, 0, 31)), ShouldBeTrue)
		})
	})
}

func TestIsSamePassword(t *testing.T) {
	info := AuthInfo{}
	info.SetPassword("secret")
	if !info.IsSamePassword("secret") {
		t.Fatalf("got AuthInfo.HashedPassword = %v, want a hashed \"secret\"", info.HashedPassword)
	}
}

func TestGetSetProviderInfoData(t *testing.T) {
	Convey("Test Get/Set ProviderInfo Data", t, func() {
		k := "com.example:johndoe"
		v := map[string]interface{}{
			"hello": "world",
		}

		Convey("Test Set ProviderInfo", func() {
			info := AuthInfo{}
			info.SetProviderInfoData(k, v)

			So(info.ProviderInfo[k], ShouldResemble, v)
		})

		Convey("Test nonexistent Get ProviderInfo", func() {
			info := AuthInfo{
				ProviderInfo: ProviderInfo{},
			}

			So(info.GetProviderInfoData(k), ShouldBeNil)
		})

		Convey("Test Get ProviderInfo", func() {
			info := AuthInfo{
				ProviderInfo: ProviderInfo(map[string]map[string]interface{}{
					k: v,
				}),
			}

			So(info.GetProviderInfoData(k), ShouldResemble, v)
		})

		Convey("Test Remove ProviderInfo", func() {
			info := AuthInfo{
				ProviderInfo: ProviderInfo(map[string]map[string]interface{}{
					k: v,
				}),
			}

			info.RemoveProviderInfoData(k)
			v, _ = info.ProviderInfo[k]
			So(v, ShouldBeNil)
		})
	})
}
