package accountmanagement

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTokenCheckUser(t *testing.T) {
	Convey("Token.CheckUser", t, func() {
		token := Token{UserID: "user"}

		err := token.CheckOAuthUser("")
		So(errors.Is(err, ErrOAuthTokenNotBoundToUser), ShouldBeTrue)

		err = token.CheckOAuthUser("user")
		So(err, ShouldBeNil)
	})
}

func TestTokenCheckState(t *testing.T) {
	Convey("Token.CheckState", t, func() {
		Convey("always succeed if token is not bound to state", func() {
			tokenNotBoundToState := Token{State: ""}

			err := tokenNotBoundToState.CheckState("")
			So(err, ShouldBeNil)

			err = tokenNotBoundToState.CheckState("state")
			So(err, ShouldBeNil)
		})

		Convey("check state", func() {
			tokenBoundToState := Token{State: "state"}

			err := tokenBoundToState.CheckState("")
			So(errors.Is(err, ErrOAuthStateNotBoundToToken), ShouldBeTrue)

			err = tokenBoundToState.CheckState("wrongstate")
			So(errors.Is(err, ErrOAuthStateNotBoundToToken), ShouldBeTrue)

			err = tokenBoundToState.CheckState("state")
			So(err, ShouldBeNil)
		})
	})
}
