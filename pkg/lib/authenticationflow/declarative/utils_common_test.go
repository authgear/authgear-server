package declarative

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestCheckLoginIDFormatForIdentification(t *testing.T) {
	Convey("checkLoginIDFormatForIdentification", t, func() {
		ctx := context.Background()
		check := func(i model.AuthenticationFlowIdentification, v string) error {
			return checkLoginIDFormatForIdentification(ctx, i, v)
		}

		Convey("phone", func() {
			So(check(model.AuthenticationFlowIdentificationPhone, "+85223456789"), ShouldBeNil)
			So(check(model.AuthenticationFlowIdentificationPhone, "user+tag@example.com"), ShouldBeError)
			So(check(model.AuthenticationFlowIdentificationPhone, "alice"), ShouldBeError)
		})

		Convey("email", func() {
			So(check(model.AuthenticationFlowIdentificationEmail, "user+tag@example.com"), ShouldBeNil)
			So(check(model.AuthenticationFlowIdentificationEmail, "+85223456789"), ShouldBeError)
			So(check(model.AuthenticationFlowIdentificationEmail, "e2e_login"), ShouldBeError)
		})

		Convey("username is not checked", func() {
			So(check(model.AuthenticationFlowIdentificationUsername, "user@example.com"), ShouldBeNil)
		})
	})
}
