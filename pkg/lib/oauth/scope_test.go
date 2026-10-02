package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestValidateScopesByClientConfig(t *testing.T) {
	Convey("ValidateScopesByClientConfig", t, func() {
		Convey("openid alone with nil knownResourceScopes is valid", func() {
			err := ValidateScopesByClientConfig([]string{"openid"}, nil)
			So(err, ShouldBeNil)
		})

		Convey("every built-in scope is known, whoever the client is", func() {
			err := ValidateScopesByClientConfig(AllowedScopes, nil)
			So(err, ShouldBeNil)
		})

		Convey("an unknown scope is invalid_scope", func() {
			err := ValidateScopesByClientConfig([]string{"openid", "foobar"}, nil)
			So(err, ShouldBeError, "unknown scope: foobar")
		})

		Convey("a resource-specific scope in knownResourceScopes is valid", func() {
			err := ValidateScopesByClientConfig([]string{"openid", "read:orders"}, []string{"read:orders"})
			So(err, ShouldBeNil)
		})

		Convey("a resource-specific scope with no resource requested is invalid_scope", func() {
			err := ValidateScopesByClientConfig([]string{"openid", "read:orders"}, nil)
			So(err, ShouldBeError, "unknown scope: read:orders")
		})

		Convey("a resource-specific scope not defined by the requested resource is invalid_scope", func() {
			err := ValidateScopesByClientConfig([]string{"openid", "read:orders"}, []string{"read:inventory"})
			So(err, ShouldBeError, "unknown scope: read:orders")
		})

		Convey("missing openid is invalid_scope", func() {
			err := ValidateScopesByClientConfig([]string{"profile"}, nil)
			So(err, ShouldBeError, "must request 'openid' scope")
		})

		Convey("pre-authenticated-url without device_sso is invalid_scope", func() {
			err := ValidateScopesByClientConfig([]string{"openid", PreAuthenticatedURLScope}, nil)
			So(err, ShouldBeError, "device_sso must be requested when using pre-authenticated url")
		})
	})
}

func TestGrantedScopes(t *testing.T) {
	Convey("GrantedScopes", t, func() {
		spa := &config.OAuthClientConfig{
			ApplicationType:                config.OAuthClientApplicationTypeSPA,
			GrantTypes_do_not_use_directly: []string{"authorization_code", "refresh_token"},
		}

		Convey("keeps every scope the client is granted, in request order", func() {
			scopes := []string{"profile", "openid", OfflineAccess, FullAccessScope, FullUserInfoScope}
			So(GrantedScopes(spa, scopes, nil), ShouldResemble, scopes)
		})

		Convey("keeps offline_access without refresh_token in grant_types, since it is always allowed", func() {
			client := &config.OAuthClientConfig{ApplicationType: config.OAuthClientApplicationTypeSPA}
			So(GrantedScopes(client, []string{"openid", OfflineAccess}, nil), ShouldResemble, []string{"openid", OfflineAccess})
		})

		Convey("drops full-access for a confidential client", func() {
			client := &config.OAuthClientConfig{ApplicationType: config.OAuthClientApplicationTypeConfidential}
			So(GrantedScopes(client, []string{"openid", FullAccessScope}, nil), ShouldResemble, []string{"openid"})
		})

		Convey("drops full-access and full-userinfo for third-party clients", func() {
			for _, typ := range []config.OAuthClientApplicationType{
				config.OAuthClientApplicationTypeThirdPartyApp,
				config.OAuthClientApplicationTypeDynamicThirdParty,
			} {
				client := &config.OAuthClientConfig{ApplicationType: typ}
				So(GrantedScopes(client, []string{"openid", "email", FullAccessScope, FullUserInfoScope}, nil), ShouldResemble, []string{"openid", "email"})
			}
		})

		Convey("drops device_sso and pre-authenticated-url unless pre-authenticated URL is enabled", func() {
			scopes := []string{"openid", DeviceSSOScope, PreAuthenticatedURLScope}
			So(GrantedScopes(spa, scopes, nil), ShouldResemble, []string{"openid"})

			enabled := &config.OAuthClientConfig{
				ApplicationType:            config.OAuthClientApplicationTypeSPA,
				PreAuthenticatedURLEnabled: true,
			}
			So(GrantedScopes(enabled, scopes, nil), ShouldResemble, scopes)
		})

		Convey("drops every built-in scope for an m2m client", func() {
			client := &config.OAuthClientConfig{ApplicationType: config.OAuthClientApplicationTypeM2M}
			So(GrantedScopes(client, []string{"openid", "profile", "read:orders"}, []string{"read:orders"}), ShouldResemble, []string{"read:orders"})
		})

		Convey("keeps a resource-specific scope only when granted", func() {
			So(GrantedScopes(spa, []string{"openid", "read:orders", "delete:orders"}, []string{"read:orders"}), ShouldResemble, []string{"openid", "read:orders"})
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

func TestIsResourceScope(t *testing.T) {
	Convey("IsResourceScope", t, func() {
		Convey("false for every member of AllowedScopes", func() {
			for _, s := range AllowedScopes {
				So(IsResourceScope(s), ShouldBeFalse)
			}
		})

		Convey("true for a resource-specific scope", func() {
			So(IsResourceScope("read:orders"), ShouldBeTrue)
		})
	})
}
