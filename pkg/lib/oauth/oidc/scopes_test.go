package oidc

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

func TestScopeAllowsClaim(t *testing.T) {
	Convey("ScopeAllowsClaim", t, func() {
		Convey("full access scope allows everything", func() {
			scope := oauth.FullAccessScope

			So(ScopeAllowsClaim(scope, ""), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "foobar"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "family_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "given_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "middle_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "nickname"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "preferred_username"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "profile"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "picture"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "website"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "gender"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "birthdate"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "zoneinfo"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "locale"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "updated_at"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "email"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "email_verified"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "address"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "phone_number"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "phone_number_verified"), ShouldBeTrue)
		})

		Convey("full user info scope allows everything", func() {
			scope := oauth.FullUserInfoScope

			So(ScopeAllowsClaim(scope, ""), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "foobar"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "family_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "given_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "middle_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "nickname"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "preferred_username"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "profile"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "picture"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "website"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "gender"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "birthdate"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "zoneinfo"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "locale"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "updated_at"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "email"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "email_verified"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "address"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "phone_number"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "phone_number_verified"), ShouldBeTrue)
		})

		Convey("profile scope allows the claims specified in the spec", func() {
			scope := ScopeProfile

			So(ScopeAllowsClaim(scope, ""), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "foobar"), ShouldBeFalse)

			So(ScopeAllowsClaim(scope, "name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "family_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "given_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "middle_name"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "nickname"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "preferred_username"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "profile"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "picture"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "website"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "gender"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "birthdate"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "zoneinfo"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "locale"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "updated_at"), ShouldBeTrue)

			So(ScopeAllowsClaim(scope, "email"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "email_verified"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "address"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "phone_number"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "phone_number_verified"), ShouldBeFalse)
		})

		Convey("email scope allows email and email_verified", func() {
			scope := ScopeEmail

			So(ScopeAllowsClaim(scope, ""), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "foobar"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "family_name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "given_name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "middle_name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "nickname"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "preferred_username"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "profile"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "picture"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "website"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "gender"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "birthdate"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "zoneinfo"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "locale"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "updated_at"), ShouldBeFalse)

			So(ScopeAllowsClaim(scope, "email"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "email_verified"), ShouldBeTrue)

			So(ScopeAllowsClaim(scope, "address"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "phone_number"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "phone_number_verified"), ShouldBeFalse)
		})

		Convey("phone scope allows phone_number and phone_number_verified", func() {
			scope := ScopePhone

			So(ScopeAllowsClaim(scope, ""), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "foobar"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "family_name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "given_name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "middle_name"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "nickname"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "preferred_username"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "profile"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "picture"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "website"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "gender"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "birthdate"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "zoneinfo"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "locale"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "updated_at"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "email"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "email_verified"), ShouldBeFalse)
			So(ScopeAllowsClaim(scope, "address"), ShouldBeFalse)

			So(ScopeAllowsClaim(scope, "phone_number"), ShouldBeTrue)
			So(ScopeAllowsClaim(scope, "phone_number_verified"), ShouldBeTrue)
		})
	})
}
