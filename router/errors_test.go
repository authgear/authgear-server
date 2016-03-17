// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"errors"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/skyerr"
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
