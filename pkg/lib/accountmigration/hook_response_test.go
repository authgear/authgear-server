package accountmigration_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	am "github.com/authgear/authgear-server/pkg/lib/accountmigration"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func TestParseHookResponse(t *testing.T) {
	Convey("ParseHookResponse", t, func() {
		ctx := context.Background()
		pass := func(raw string, expected *am.HookResponse) {
			r := strings.NewReader(raw)
			actual, err := am.ParseHookResponse(ctx, r)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		fail := func(raw string, errString string) {
			r := strings.NewReader(raw)
			_, err := am.ParseHookResponse(ctx, r)
			So(err, ShouldBeError, errString)
		}

		pass(`
		{
			"identities": [
				{
					"type": "login_id",
					"login_id": {
						"key": "email",
						"type": "email",
						"value": "faseng@example.com"
					}
				}
			]
		}
		`, &am.HookResponse{
			Identities: []*identity.MigrateSpec{
				{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDMigrateSpec{
						Key:   "email",
						Type:  "email",
						Value: "faseng@example.com",
					},
				},
			},
		})

		pass(`
		{
			"identities": [
				{
					"type": "login_id",
					"login_id": {
						"key": "email",
						"type": "email",
						"value": "faseng@example.com"
					}
				}
			],
			"authenticators": [
				{
					"type": "oob_otp_email",
					"oobotp": {
						"email": "faseng@example.com"
					}
				}
			]
		}
		`, &am.HookResponse{
			Identities: []*identity.MigrateSpec{
				{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDMigrateSpec{
						Key:   "email",
						Type:  "email",
						Value: "faseng@example.com",
					},
				},
			},
			Authenticators: []*authenticator.MigrateSpec{
				{
					Type: model.AuthenticatorTypeOOBEmail,
					OOBOTP: &authenticator.OOBOTPMigrateSpec{
						Email: "faseng@example.com",
					},
				},
			},
		})

		fail(`
		{
			"identities": [
				{
					"type": "login_id",
					"login_id": {
						"key": "email",
						"type": "email",
						"value": "faseng@example.com"
					}
			  	}
			],
			"authenticators": [
				{
					"type": "oob_otp_sms",
					"oobotp": {}
				}
			]
		}
		`, `invalid value:
/authenticators/0/oobotp: required
  map[actual:<nil> expected:[phone] missing:[phone]]`)

		fail(`
		{
			"identities": [
				{
					"type": "login_id",
					"login_id": {
						"key": "email",
						"type": "email"
					}
			  	}
			]
		}
		`, `invalid value:
/identities/0/login_id: required
  map[actual:[key type] expected:[key type value] missing:[value]]`)

		fail(`{}`, `invalid value:
<root>: required
  map[actual:<nil> expected:[identities] missing:[identities]]`)

	})
}
