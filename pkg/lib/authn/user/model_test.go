package user

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func TestComputeUserEndUserActionID(t *testing.T) {

	// Convey("EndUserAccountID", t, func() {
	// So((&User{}).EndUserAccountID(), ShouldEqual, "")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "email": "user@example.com",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user@example.com")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "preferred_username": "user",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "phone_number": "+85298765432",
	// },
	// }).EndUserAccountID(), ShouldEqual, "+85298765432")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "preferred_username": "user",
	// "phone_number":       "+85298765432",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "email":              "user@example.com",
	// "preferred_username": "user",
	// "phone_number":       "+85298765432",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user@example.com")
	// })
	Convey("ComputeUserEndUserActionID", t, func() {
		So(computeEndUserAccountID(map[string]interface{}{}, nil), ShouldEqual, "")

		So(computeEndUserAccountID(
			map[string]interface{}{
				"email": "user@example.com",
			},
			nil), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(
			map[string]interface{}{
				"preferred_username": "user",
			},
			nil), ShouldEqual, "user")

		So(computeEndUserAccountID(
			map[string]interface{}{
				"phone_number": "+85298765432",
			},
			nil), ShouldEqual, "+85298765432")

		So(computeEndUserAccountID(
			map[string]interface{}{
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
			nil), ShouldEqual, "user")

		So(computeEndUserAccountID(
			map[string]interface{}{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
			nil), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(
			map[string]interface{}{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
			[]*identity.Info{
				{
					Type: model.IdentityTypeLDAP,
					LDAP: &identity.LDAP{
						RawEntryJSON: map[string]interface{}{
							"dn": "cn=user,dc=example,dc=org",
						},
					},
				},
			}), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(map[string]interface{}{}, []*identity.Info{
			{
				Type: model.IdentityTypeLDAP,
				LDAP: &identity.LDAP{
					RawEntryJSON: map[string]interface{}{
						"dn": "cn=user,dc=example,dc=org",
					},
				},
			},
		}), ShouldEqual, "cn=user,dc=example,dc=org")

		So(computeEndUserAccountID(map[string]interface{}{}, []*identity.Info{
			{
				Type: model.IdentityTypeLDAP,
				LDAP: &identity.LDAP{
					UserIDAttributeName:  "uid",
					UserIDAttributeValue: []byte("example-user"),
				},
			},
		}), ShouldEqual, "uid=example-user")
	})
}
