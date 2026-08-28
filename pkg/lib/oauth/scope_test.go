package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestValidateScopesByClientConfig(t *testing.T) {
	Convey("ValidateScopesByClientConfig", t, func() {
		client := &config.OAuthClientConfig{
			ClientID:                       "client",
			ApplicationType:                config.OAuthClientApplicationTypeSPA,
			GrantTypes_do_not_use_directly: []string{"authorization_code", "refresh_token"},
		}

		Convey("baseline: openid alone with nil allowedResourceScopes is unchanged", func() {
			err := ValidateScopesByClientConfig(client, []string{"openid"}, nil)
			So(err, ShouldBeNil)
		})

		Convey("a resource-specific scope matching allowedResourceScopes is allowed", func() {
			err := ValidateScopesByClientConfig(client, []string{"openid", "read:orders"}, []string{"read:orders"})
			So(err, ShouldBeNil)
		})

		Convey("a resource-specific scope with no matching resource (nil allowedResourceScopes) is invalid_scope", func() {
			err := ValidateScopesByClientConfig(client, []string{"openid", "read:orders"}, nil)
			So(err, ShouldBeError, "specified scope is not allowed: read:orders")
		})

		Convey("a resource-specific scope not in the resource's own scope list is invalid_scope", func() {
			err := ValidateScopesByClientConfig(client, []string{"openid", "read:orders"}, []string{"read:inventory"})
			So(err, ShouldBeError, "specified scope is not allowed: read:orders")
		})
	})
}

func TestScopeAllowsClaim(t *testing.T) {
	Convey("ScopeAllowsClaim", t, func() {
		Convey("full access scope allows everything", func() {
			scope := FullAccessScope

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
			scope := FullUserInfoScope

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
