package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

func TestGithub(t *testing.T) {
	Convey("Github", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
		}
		g := Github{}

		ctx := context.Background()
		u, err := g.GetAuthorizationURL(ctx, deps, oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI: "https://localhost/",
			Nonce:       "nonce",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://github.com/login/oauth/authorize?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&scope=read%3Auser+user%3Aemail&state=state")
	})
}

func TestGithubGetUserProfile(t *testing.T) {
	Convey("Github GetUserProfile", t, func() {
		client := &http.Client{}
		gock.InterceptClient(client)
		defer gock.Off()
		defer gock.RestoreClient(client)
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
			ClientSecret: "client_secret",
			HTTPClient:   client,
		}
		g := Github{}
		g.SetDefaults(deps.ProviderConfig)
		options := oauthrelyingparty.GetUserProfileOptions{
			Query:       "code=code",
			RedirectURI: "https://localhost/",
		}
		mockUserInfo := func() {
			gock.New("https://api.github.com").Get("/user").
				MatchHeader("Authorization", "Bearer access_token").
				Reply(http.StatusOK).
				BodyString(`{"id":123,"login":"octocat","email":"octocat@example.com"}`)
		}

		for _, status := range []int{http.StatusOK, http.StatusBadRequest} {
			Convey(http.StatusText(status), func() {
				Convey("returns the OAuth error without fetching user info", func() {
					gock.New(githubTokenURL).Post("").Reply(status).
						BodyString(`{"error":"bad_verification_code","error_description":"The code passed is incorrect or expired.","error_uri":"https://docs.github.com/apps"}`)
					mockUserInfo()

					profile, err := g.GetUserProfile(context.Background(), deps, options)
					So(err, ShouldResemble, &oauthrelyingparty.ErrorResponse{
						Error_:           "bad_verification_code",
						ErrorDescription: "The code passed is incorrect or expired.",
						ErrorURI:         "https://docs.github.com/apps",
					})
					So(profile, ShouldResemble, oauthrelyingparty.UserProfile{})
					So(gock.Pending(), ShouldHaveLength, 1)
				})

				Convey("propagates malformed JSON without fetching user info", func() {
					gock.New(githubTokenURL).Post("").Reply(status).
						BodyString(`{"error":`)
					mockUserInfo()

					_, err := g.GetUserProfile(context.Background(), deps, options)
					So(err, ShouldNotBeNil)
					So(gock.Pending(), ShouldHaveLength, 1)
				})
			})
		}

		Convey("fetches user info after a successful code exchange", func() {
			gock.New(githubTokenURL).Post("").
				MatchHeader("Accept", "application/json").
				MatchHeader("Content-Type", "application/x-www-form-urlencoded").
				BodyString("client_id=client_id&client_secret=client_secret&code=code&redirect_uri=https%3A%2F%2Flocalhost%2F").
				Reply(http.StatusOK).
				BodyString(`{"access_token":"access_token","token_type":"bearer","scope":"read:user,user:email"}`)
			mockUserInfo()

			profile, err := g.GetUserProfile(context.Background(), deps, options)
			So(err, ShouldBeNil)
			So(profile.ProviderUserID, ShouldEqual, "123")
			So(profile.ProviderRawProfile["id"], ShouldResemble, json.Number("123"))
			So(profile.StandardAttributes[stdattrs.Email], ShouldEqual, "octocat@example.com")
			So(gock.IsDone(), ShouldBeTrue)
		})
	})
}
