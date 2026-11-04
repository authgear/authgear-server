package identity

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestInfoJSON(t *testing.T) {
	Convey("Info JSON", t, func() {
		test := func(i *Info) {
			bytes, err := json.Marshal(i)
			So(err, ShouldBeNil)

			var ii Info
			err = json.Unmarshal(bytes, &ii)
			So(err, ShouldBeNil)

			So(i, ShouldResemble, &ii)
		}

		test(&Info{
			ID:        "id",
			UserID:    "userid",
			CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:      model.IdentityTypeLoginID,

			LoginID: &LoginID{
				ID:        "id",
				UserID:    "userid",
				CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),

				LoginIDKey:      "email",
				LoginIDType:     model.LoginIDKeyTypeEmail,
				LoginID:         "user@example.com",
				OriginalLoginID: "user@example.com",
				UniqueKey:       "user@example.com",
				Claims: map[string]any{
					"email": "user@example.com",
				},
			},
		})

		test(&Info{
			ID:        "id",
			UserID:    "userid",
			CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:      model.IdentityTypeOAuth,

			OAuth: &OAuth{
				ID:        "id",
				UserID:    "userid",
				CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),

				ProviderID: oauthrelyingparty.ProviderID{
					Type: "provider",
					Keys: map[string]any{
						"client_id": "client_id",
					},
				},
				ProviderSubjectID: "sub",
				UserProfile: map[string]any{
					"email": "user@example.com",
				},
				Claims: map[string]any{
					"email": "user@example.com",
				},
			},
		})

		test(&Info{
			ID:        "id",
			UserID:    "userid",
			CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:      model.IdentityTypeAnonymous,

			Anonymous: &Anonymous{
				ID:        "id",
				UserID:    "userid",
				CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),

				KeyID: "keyid",
				Key:   []byte("abc"),
			},
		})

		test(&Info{
			ID:        "id",
			UserID:    "userid",
			CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:      model.IdentityTypeBiometric,

			Biometric: &Biometric{
				ID:        "id",
				UserID:    "userid",
				CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),

				KeyID: "keyid",
				Key:   []byte("abc"),
				DeviceInfo: map[string]any{
					"name": "name",
				},
			},
		})

		test(&Info{
			ID:        "id",
			UserID:    "userid",
			CreatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			Type:      model.IdentityTypePasskey,

			Passkey: &Passkey{
				ID:           "id",
				UserID:       "userid",
				CreatedAt:    time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				CredentialID: "credentialid",
				CreationOptions: &model.WebAuthnCreationOptions{
					PublicKey: model.PublicKeyCredentialCreationOptions{},
				},
				AttestationResponse: []byte("abc"),
				AAGUID:              "fbfc3007-154e-4ecc-8c0b-6e020557d7bd",
			},
		})
	})
}

func TestInfoToModelPasskey(t *testing.T) {
	Convey("Info.ToModel of a passkey", t, func() {
		makeInfo := func(aaguid string) *Info {
			return (&Passkey{
				ID:           "id",
				UserID:       "userid",
				CredentialID: "credentialid",
				CreationOptions: &model.WebAuthnCreationOptions{
					PublicKey: model.PublicKeyCredentialCreationOptions{
						User: model.PublicKeyCredentialUserEntity{
							DisplayName: "user@example.com",
						},
					},
				},
				AAGUID: aaguid,
			}).ToInfo()
		}

		Convey("has the AAGUID and the provider name of a known AAGUID", func() {
			claims := makeInfo("fbfc3007-154e-4ecc-8c0b-6e020557d7bd").ToModel().Claims
			So(claims[IdentityClaimPasskeyAAGUID], ShouldEqual, "fbfc3007-154e-4ecc-8c0b-6e020557d7bd")
			So(claims[IdentityClaimPasskeyProviderName], ShouldEqual, "Apple Passwords")
		})

		Convey("has no provider name for the all-zero AAGUID", func() {
			claims := makeInfo("00000000-0000-0000-0000-000000000000").ToModel().Claims
			So(claims[IdentityClaimPasskeyAAGUID], ShouldEqual, "00000000-0000-0000-0000-000000000000")
			So(claims, ShouldNotContainKey, IdentityClaimPasskeyProviderName)
		})

		Convey("has neither claim without an AAGUID", func() {
			claims := makeInfo("").ToModel().Claims
			So(claims, ShouldNotContainKey, IdentityClaimPasskeyAAGUID)
			So(claims, ShouldNotContainKey, IdentityClaimPasskeyProviderName)
		})
	})
}
