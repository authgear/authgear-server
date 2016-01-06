package router

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"

	"github.com/oursky/skygear/skyerr"
)

func TestErrors(t *testing.T) {
	Convey("defaultStatusCode", t, func() {
		Convey("not authenticated as unauthorized", func() {
			httpStatus := defaultStatusCode(skyerr.NewError(skyerr.NotAuthenticated, "an error"))
			So(httpStatus, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("undefined code as internal server error", func() {
			httpStatus := defaultStatusCode(skyerr.NewError(999, "an error"))
			So(httpStatus, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("unexpected error as internal server error", func() {
			httpStatus := defaultStatusCode(skyerr.NewError(skyerr.UnexpectedError, "an error"))
			So(httpStatus, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("errorFromRecoveringPanic", t, func() {
		Convey("return original skyerr", func() {
			err := skyerr.NewError(skyerr.InvalidArgument, "an error")
			newError := errorFromRecoveringPanic(err)
			So(newError, ShouldResemble, err)
		})

		Convey("wrap error with skyerror", func() {
			err := errors.New("an error")
			newError := errorFromRecoveringPanic(err)
			So(newError, ShouldResemble, skyerr.NewErrorf(skyerr.UnexpectedError, "panic occurred while handling request: an error"))
		})

		Convey("wrap unexpected type with skyerror", func() {
			err := "an error"
			newError := errorFromRecoveringPanic(err)
			So(newError, ShouldResemble, skyerr.NewErrorf(skyerr.UnexpectedError, "an panic occurred and the error is not known"))
		})
	})
}
