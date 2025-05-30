package identity

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func TestOAuthSpec_ToClaimsForIDToken(t *testing.T) {
	Convey("TestOAuthSpec_ToClaimsForIDToken", t, func() {
		Convey("should return correct claims when RawProfile is nil", func() {
			s := OAuthSpec{
				ProviderAlias: "github",
				ProviderID: oauthrelyingparty.ProviderID{
					Type: "oauth",
					Keys: map[string]any{
						"client_id": "test_client_id",
					},
				},
				SubjectID:  "test_subject_id",
				RawProfile: nil,
			}

			claims := s.ToClaimsForIDToken()

			So(claims, ShouldNotBeNil)
			So(claims["https://authgear.com/claims/oauth/provider_alias"], ShouldEqual, "github")
			So(claims["https://authgear.com/claims/oauth/provider_type"], ShouldEqual, "oauth")
			So(claims["https://authgear.com/claims/oauth/subject_id"], ShouldEqual, "test_subject_id")
			So(claims["https://authgear.com/claims/oauth/profile"], ShouldEqual, make(map[string]any))
		})

		Convey("should return correct claims when RawProfile is not nil", func() {
			rawProfile := map[string]any{"email": "test@example.com"}
			s := OAuthSpec{
				ProviderAlias: "google",
				ProviderID: oauthrelyingparty.ProviderID{
					Type: "google",
					Keys: map[string]any{
						"client_id": "test_client_id_google",
					},
				},
				SubjectID:  "test_subject_id_google",
				RawProfile: rawProfile,
			}

			claims := s.ToClaimsForIDToken()

			So(claims, ShouldNotBeNil)
			So(claims["https://authgear.com/claims/oauth/provider_alias"], ShouldEqual, "google")
			So(claims["https://authgear.com/claims/oauth/provider_type"], ShouldEqual, "google")
			So(claims["https://authgear.com/claims/oauth/subject_id"], ShouldEqual, "test_subject_id_google")
			So(claims["https://authgear.com/claims/oauth/profile"], ShouldResemble, rawProfile)
		})
	})
}
