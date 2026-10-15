package stdattrs

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestServiceNoEventValidateStringConstraints(t *testing.T) {
	Convey("ServiceNoEvent.validateStringConstraints", t, func() {
		two := 2
		service := &ServiceNoEvent{
			UserProfileConfig: &config.UserProfileConfig{
				StandardAttributes: &config.StandardAttributesConfig{
					Validation: []*config.StandardAttributesValidationConfig{
						{Pointer: "/given_name", MinLength: &two},
						{Pointer: "/address/locality", Pattern: "^[A-Z]"},
					},
				},
			},
		}

		Convey("checks a changed value", func() {
			err := service.validateStringConstraints(map[string]any{}, map[string]any{"given_name": "A"})
			var aErr *validation.AggregatedError
			So(errors.As(err, &aErr), ShouldBeTrue)
			So(aErr.Errors, ShouldHaveLength, 1)
			So(aErr.Errors[0].Keyword, ShouldEqual, "minLength")
			So(aErr.Errors[0].Location, ShouldEqual, "/given_name")

			So(service.validateStringConstraints(map[string]any{}, map[string]any{"given_name": "Al"}), ShouldBeNil)
		})

		Convey("passes an unchanged violating value", func() {
			err := service.validateStringConstraints(
				map[string]any{"given_name": "A"},
				map[string]any{"given_name": "A"},
			)
			So(err, ShouldBeNil)
		})

		Convey("reports a nested location", func() {
			err := service.validateStringConstraints(
				map[string]any{},
				map[string]any{"address": map[string]any{"locality": "london"}},
			)
			var aErr *validation.AggregatedError
			So(errors.As(err, &aErr), ShouldBeTrue)
			So(aErr.Errors[0].Keyword, ShouldEqual, "pattern")
			So(aErr.Errors[0].Location, ShouldEqual, "/address/locality")
		})

		Convey("ignores absent values", func() {
			So(service.validateStringConstraints(map[string]any{"given_name": "A"}, map[string]any{}), ShouldBeNil)
		})
	})
}

func TestServiceNoEventDeriveStandardAttributesForUsersWithClaims(t *testing.T) {
	Convey("ServiceNoEvent.DeriveStandardAttributesForUsersWithClaims", t, func() {
		service := &ServiceNoEvent{
			UserProfileConfig: &config.UserProfileConfig{
				StandardAttributes: &config.StandardAttributesConfig{},
			},
			Transformer: &PictureTransformer{},
		}

		updatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		claims := []*verification.Claim{
			{UserID: "user-id", Name: "email", Value: "foo@example.com"},
			{UserID: "user-id", Name: "phone_number", Value: "+85212345678"},
			{UserID: "user-id", Name: "some_unrelated_claim", Value: "irrelevant"},
		}

		Convey("sets email_verified and phone_number_verified when the values match a verified claim", func() {
			result, err := service.DeriveStandardAttributesForUsersWithClaims(
				context.Background(),
				accesscontrol.RoleGreatest,
				[]string{"user-id"},
				[]time.Time{updatedAt},
				[]map[string]any{{
					"email":        "foo@example.com",
					"phone_number": "+85212345678",
				}},
				claims,
			)
			So(err, ShouldBeNil)
			So(result["user-id"]["email"], ShouldEqual, "foo@example.com")
			So(result["user-id"]["email_verified"], ShouldEqual, true)
			So(result["user-id"]["phone_number"], ShouldEqual, "+85212345678")
			So(result["user-id"]["phone_number_verified"], ShouldEqual, true)
		})

		Convey("sets email_verified and phone_number_verified to false when the values do not match a verified claim", func() {
			result, err := service.DeriveStandardAttributesForUsersWithClaims(
				context.Background(),
				accesscontrol.RoleGreatest,
				[]string{"user-id"},
				[]time.Time{updatedAt},
				[]map[string]any{{
					"email":        "bar@example.com",
					"phone_number": "+85287654321",
				}},
				claims,
			)
			So(err, ShouldBeNil)
			So(result["user-id"]["email_verified"], ShouldEqual, false)
			So(result["user-id"]["phone_number_verified"], ShouldEqual, false)
		})

		Convey("ignores claims of unrelated names", func() {
			result, err := service.DeriveStandardAttributesForUsersWithClaims(
				context.Background(),
				accesscontrol.RoleGreatest,
				[]string{"user-id"},
				[]time.Time{updatedAt},
				[]map[string]any{{
					"some_unrelated_claim": "irrelevant",
				}},
				claims,
			)
			So(err, ShouldBeNil)
			So(result["user-id"], ShouldContainKey, "some_unrelated_claim")
			So(result["user-id"], ShouldNotContainKey, "some_unrelated_claim_verified")
		})
	})
}
