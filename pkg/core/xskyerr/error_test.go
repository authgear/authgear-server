package skyerr_test

import (
	"encoding/json"
	"testing"

	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAPIError(t *testing.T) {
	Convey("AsAPIError", t, func() {
		Convey("simple error", func() {
			err := skyerr.NewInternalError("internal server error")
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.InternalError, Reason: string(skyerr.InternalError)},
				Message: "internal server error",
				Code:    500,
				Info:    map[string]interface{}{},
			})
		})
		Convey("common error", func() {
			NotAuthenticated := skyerr.Unauthorized.WithReason("NotAuthenticated")
			err := NotAuthenticated.New("authentication required")
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Unauthorized, Reason: "NotAuthenticated"},
				Message: "authentication required",
				Code:    401,
				Info:    map[string]interface{}{},
			})
		})
		Convey("error with details", func() {
			NotAuthenticated := skyerr.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithDetails(
				"failed to validate form payload",
				skyerr.Details{
					"field":   skyerr.APIErrorString("email"),
					"user_id": "user-id",
				},
			)
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info: map[string]interface{}{
					"field": skyerr.APIErrorString("email"),
				},
			})
		})
	})

	Convey("APIError", t, func() {
		Convey("simple error", func() {
			apiErr := &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.InternalError, Reason: string(skyerr.InternalError)},
				Message: "internal server error",
				Code:    500,
				Info:    map[string]interface{}{},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"InternalError","reason":"InternalError","message":"internal server error","code":500}`)
		})
		Convey("common error", func() {
			apiErr := &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Unauthorized, Reason: "NotAuthenticated"},
				Message: "authentication required",
				Code:    401,
				Info:    map[string]interface{}{},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Unauthorized","reason":"NotAuthenticated","message":"authentication required","code":401}`)
		})
		Convey("error with details", func() {
			apiErr := &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info: map[string]interface{}{
					"field": skyerr.APIErrorString("email"),
				},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Invalid","reason":"ValidationFailure","message":"failed to validate form payload","code":400,"info":{"field":"email"}}`)
		})
	})
}
