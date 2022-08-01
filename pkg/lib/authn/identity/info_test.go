package identity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
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
				LoginIDType:     config.LoginIDKeyTypeEmail,
				LoginID:         "user@example.com",
				OriginalLoginID: "user@example.com",
				UniqueKey:       "user@example.com",
				Claims: map[string]interface{}{
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

				ProviderID: config.ProviderID{
					Type: "provider",
					Keys: map[string]interface{}{
						"client_id": "client_id",
					},
				},
				ProviderSubjectID: "sub",
				UserProfile: map[string]interface{}{
					"email": "user@example.com",
				},
				Claims: map[string]interface{}{
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
				DeviceInfo: map[string]interface{}{
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
			},
		})
	})
}
