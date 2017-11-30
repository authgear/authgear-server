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

package pq

import (
	"database/sql"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOAuthCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		now := time.Now()
		oauthinfo := skydb.OAuthInfo{
			UserID:          "userid",
			Provider:        "skygear",
			PrincipalID:     "faseng",
			TokenResponse:   map[string]interface{}{"access_token": "token"},
			ProviderProfile: map[string]interface{}{"name": "faseng cto"},
			CreatedAt:       &now,
			UpdatedAt:       &now,
		}

		Convey("creates oauth record", func() {
			err := c.CreateOAuthInfo(&oauthinfo)
			So(err, ShouldBeNil)

			provider := []byte{}
			tokenResponse := tokenResponseValue{}
			providerProfile := providerProfileValue{}
			err = c.QueryRowx("SELECT provider, token_response, profile FROM _sso_oauth WHERE user_id = 'userid'").
				Scan(&provider, &tokenResponse, &providerProfile)
			So(err, ShouldBeNil)
			So(provider, ShouldResemble, []byte("skygear"))
			So(tokenResponse.TokenResponse, ShouldResemble, skydb.TokenResponse{
				"access_token": "token",
			})
			So(providerProfile.ProviderProfile, ShouldResemble, skydb.ProviderProfile{
				"name": "faseng cto",
			})
		})

		Convey("returns ErrUserDuplicated when create duplicated oauth info", func() {
			So(c.CreateOAuthInfo(&oauthinfo), ShouldBeNil)
			So(c.CreateOAuthInfo(&oauthinfo), ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("gets ouath info", func() {
			So(c.CreateOAuthInfo(&oauthinfo), ShouldBeNil)

			fetchedoauthinfo := skydb.OAuthInfo{}
			err := c.GetOAuthInfo("skygear", "faseng", &fetchedoauthinfo)
			So(err, ShouldBeNil)

			So(fetchedoauthinfo.UserID, ShouldResemble, "userid")
			So(fetchedoauthinfo.ProviderProfile, ShouldResemble, oauthinfo.ProviderProfile)
		})

		Convey("gets ouath info by provider and user id", func() {
			So(c.CreateOAuthInfo(&oauthinfo), ShouldBeNil)

			fetchedoauthinfo := skydb.OAuthInfo{}
			err := c.GetOAuthInfoByProviderAndUserID("skygear", "userid", &fetchedoauthinfo)
			So(err, ShouldBeNil)

			So(fetchedoauthinfo.PrincipalID, ShouldResemble, "faseng")
			So(fetchedoauthinfo.ProviderProfile, ShouldResemble, oauthinfo.ProviderProfile)
		})

		Convey("updates oauth info", func() {
			So(c.CreateOAuthInfo(&oauthinfo), ShouldBeNil)

			oauthinfo.TokenResponse = skydb.TokenResponse{
				"access_token": "new_token",
			}

			err := c.UpdateOAuthInfo(&oauthinfo)
			So(err, ShouldBeNil)

			tokenResponse := tokenResponseValue{}
			err = c.QueryRowx("SELECT token_response FROM _sso_oauth WHERE user_id = 'userid'").
				Scan(&tokenResponse)
			So(err, ShouldBeNil)
			So(tokenResponse.TokenResponse, ShouldResemble, skydb.TokenResponse{
				"access_token": "new_token",
			})
		})

		Convey("delete oauth info", func() {
			So(c.CreateOAuthInfo(&oauthinfo), ShouldBeNil)
			So(c.DeleteOAuth("skygear", "faseng"), ShouldBeNil)

			placeholder := []byte{}
			err := c.QueryRowx("SELECT false FROM _sso_oauth WHERE user_id = $1", "userid").Scan(&placeholder)
			So(err, ShouldEqual, sql.ErrNoRows)
			So(placeholder, ShouldBeEmpty)
		})

	})
}
