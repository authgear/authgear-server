package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestCorrectErrorCausePointer(t *testing.T) {
	Convey("Test correctErrorCausePointer", t, func() {

		Convey("should update error pointer", func() {
			err := correctErrorCausePointer(validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
				Kind:    validation.ErrorStringFormat,
				Pointer: "/0/value",
				Message: "invalid login ID format",
				Details: map[string]interface{}{"format": "email"},
			}}), 1)
			causes := validation.ErrorCauses(err)

			So(len(causes), ShouldEqual, 1)
			So(causes[0].Pointer, ShouldEqual, "/login_ids/1/value")
		})
	})
}
