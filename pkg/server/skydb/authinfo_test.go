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

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestNewAuthInfo(t *testing.T) {
	info := NewAuthInfo("userinfoid", "john.doe@example.com", "secret")

	if info.Username != "userinfoid" {
		t.Fatalf("got info.ID = %v, want userinfoid", info.ID)
	}

	if info.Email != "john.doe@example.com" {
		t.Fatalf("got info.Email = %v, want john.doe@example.com", info.Email)
	}

	if bytes.Equal(info.HashedPassword, nil) {
		t.Fatalf("got info.HashPassword = %v, want non-empty value", info.HashedPassword)
	}
}

func TestNewAuthInfoWithEmptyID(t *testing.T) {
	info := NewAuthInfo("", "jane.doe@example.com", "anothersecret")

	if info.ID == "" {
		t.Fatalf("got empty info.ID, want non-empty string")
	}

	if info.Email != "jane.doe@example.com" {
		t.Fatalf("got info.Email = %v, want jane.doe@example.com", info.Email)
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

	if info.Email != "" {
		t.Fatalf("got info.Email = %v, want empty string", info.Email)
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
