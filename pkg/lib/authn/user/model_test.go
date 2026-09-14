package user

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
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
		So(computeEndUserAccountID(map[string]any{}, nil), ShouldEqual, "")

		So(computeEndUserAccountID(
			map[string]any{
				"email": "user@example.com",
			},
			nil), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(
			map[string]any{
				"preferred_username": "user",
			},
			nil), ShouldEqual, "user")

		So(computeEndUserAccountID(
			map[string]any{
				"phone_number": "+85298765432",
			},
			nil), ShouldEqual, "+85298765432")

		So(computeEndUserAccountID(
			map[string]any{
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
			nil), ShouldEqual, "user")

		So(computeEndUserAccountID(
			map[string]any{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
			nil), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(
			map[string]any{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
			[]*identity.Info{
				{
					Type: model.IdentityTypeLDAP,
					LDAP: &identity.LDAP{
						RawEntryJSON: map[string]any{
							"dn": "cn=user,dc=example,dc=org",
						},
					},
				},
			}), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(map[string]any{}, []*identity.Info{
			{
				Type: model.IdentityTypeLDAP,
				LDAP: &identity.LDAP{
					RawEntryJSON: map[string]any{
						"dn": "cn=user,dc=example,dc=org",
					},
				},
			},
		}), ShouldEqual, "cn=user,dc=example,dc=org")

		So(computeEndUserAccountID(map[string]any{}, []*identity.Info{
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

func TestSortOptionApply(t *testing.T) {
	Convey("SortOption.Apply", t, func() {
		newQuery := func() db.SelectBuilder {
			return db.NewSQLBuilderApp("public", "app-id").Select("id").From("t")
		}
		sortByLastLogin := SortOption{
			SortBy:        SortByLastLoginAt,
			SortDirection: model.SortDirectionDesc,
		}

		Convey("uses the sort key as the column name by default", func() {
			sql, _, err := sortByLastLogin.Apply(newQuery(), "").ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldContainSubstring, "ORDER BY last_login_at desc NULLS LAST")
		})

		Convey("resolves the column through SortColumns", func() {
			sql, _, err := sortByLastLogin.ApplyWithColumns(newQuery(), "", userSortColumns).ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldContainSubstring, "ORDER BY login_at desc NULLS LAST")
			So(sql, ShouldNotContainSubstring, "last_login_at")
		})

		Convey("falls back to the sort key for keys missing from SortColumns", func() {
			byCreatedAt := SortOption{
				SortBy:        SortByCreatedAt,
				SortDirection: model.SortDirectionAsc,
			}
			sql, _, err := byCreatedAt.ApplyWithColumns(newQuery(), "", SortColumns{
				SortByLastLoginAt: "login_at",
			}).ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldContainSubstring, "ORDER BY created_at asc NULLS LAST")
		})

		Convey("filters the cursor on the resolved column", func() {
			sql, args, err := sortByLastLogin.ApplyWithColumns(newQuery(), "2026-09-10T00:00:00Z", userSortColumns).ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldContainSubstring, "login_at <")
			So(sql, ShouldNotContainSubstring, "last_login_at <")
			So(args, ShouldContain, "2026-09-10T00:00:00Z")
		})
	})
}
