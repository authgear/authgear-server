package service

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func newBool(b bool) *bool {
	return &b
}

func TestProviderListCandidates(t *testing.T) {
	Convey("Provider ListCandidates", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		loginIDProvider := NewMockLoginIDIdentityProvider(ctrl)
		oauthProvider := NewMockOAuthIdentityProvider(ctrl)

		p := &Service{
			Authentication: &config.AuthenticationConfig{},
			Identity: &config.IdentityConfig{
				LoginID: &config.LoginIDConfig{},
				OAuth:   &config.OAuthSSOConfig{},
			},
			IdentityFeatureConfig: &config.IdentityFeatureConfig{
				OAuth: &config.OAuthSSOFeatureConfig{
					Providers: &config.OAuthSSOProvidersFeatureConfig{
						Google: &config.OAuthSSOProviderFeatureConfig{
							Disabled: false,
						},
					},
				},
			},
			LoginID: loginIDProvider,
			OAuth:   oauthProvider,
		}

		Convey("no candidates", func() {
			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldBeEmpty)
		})

		Convey("oauth", func() {
			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeOAuth}
			p.Identity.OAuth.Providers = []config.OAuthSSOProviderConfig{
				{
					Alias:          "google",
					Type:           "google",
					ModifyDisabled: newBool(false),
				},
			}

			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":         "",
					"type":                "oauth",
					"display_id":          "",
					"provider_type":       "google",
					"provider_alias":      "google",
					"provider_subject_id": "",
					"provider_app_type":   "",
					"modify_disabled":     false,
				},
			})
		})

		Convey("loginid", func() {
			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeLoginID}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfig{
				{
					Type:           "email",
					Key:            "email",
					ModifyDisabled: newBool(false),
				},
			}

			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":     "",
					"type":            "login_id",
					"display_id":      "",
					"login_id_type":   "email",
					"login_id_key":    "email",
					"login_id_value":  "",
					"modify_disabled": false,
				},
			})
		})

		Convey("respect authentication", func() {
			p.Identity.OAuth.Providers = []config.OAuthSSOProviderConfig{
				{
					Alias:          "google",
					Type:           "google",
					ModifyDisabled: newBool(false),
				},
			}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfig{
				{
					Type:           "email",
					Key:            "email",
					ModifyDisabled: newBool(false),
				},
			}

			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldBeEmpty)
		})

		Convey("associate login ID identity", func() {
			userID := "a"

			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeLoginID}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfig{
				{
					Type:           "email",
					Key:            "email",
					ModifyDisabled: newBool(false),
				},
			}

			loginIDProvider.EXPECT().List(userID).Return([]*identity.LoginID{
				{
					LoginIDKey:      "email",
					LoginID:         "john.doe@example.com",
					OriginalLoginID: "john.doe@example.com",
					Claims: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
			}, nil)
			oauthProvider.EXPECT().List(userID).Return(nil, nil)

			actual, err := p.ListCandidates(userID)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":     "",
					"type":            "login_id",
					"display_id":      "john.doe@example.com",
					"login_id_type":   "email",
					"login_id_key":    "email",
					"login_id_value":  "john.doe@example.com",
					"modify_disabled": false,
				},
			})
		})

		Convey("associate oauth identity", func() {
			userID := "a"

			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeOAuth}
			p.Identity.OAuth.Providers = []config.OAuthSSOProviderConfig{
				{
					Alias:          "google",
					Type:           "google",
					ModifyDisabled: newBool(false),
				},
			}

			loginIDProvider.EXPECT().List(userID).Return(nil, nil)
			oauthProvider.EXPECT().List(userID).Return([]*identity.OAuth{
				{
					ProviderID: config.ProviderID{
						Type: "google",
						Keys: map[string]interface{}{},
					},
					ProviderSubjectID: "john.doe@gmail.com",
					Claims: map[string]interface{}{
						"email": "john.doe@gmail.com",
					},
				},
			}, nil)

			actual, err := p.ListCandidates(userID)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":         "",
					"type":                "oauth",
					"display_id":          "john.doe@gmail.com",
					"provider_type":       "google",
					"provider_alias":      "google",
					"provider_subject_id": "john.doe@gmail.com",
					"provider_app_type":   "",
					"modify_disabled":     false,
				},
			})
		})
	})
}
