package skyerr_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/authgear/authgear-server/pkg/core/skyerr"

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
		Convey("wrapped error", func() {
			var err error
			err = skyerr.NewInternalError("internal server error")
			err = fmt.Errorf("wrap this: %w", err)

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
					"field":   skyerr.APIErrorDetail.Value("email"),
					"user_id": "user-id",
				},
			)
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info: map[string]interface{}{
					"field": "email",
				},
			})
		})
		Convey("error with info", func() {
			NotAuthenticated := skyerr.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithInfo(
				"failed to validate form payload",
				skyerr.Details{"field": "email"},
			)
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info: map[string]interface{}{
					"field": "email",
				},
			})
		})
		Convey("error with cause", func() {
			NotAuthenticated := skyerr.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithCause(
				"invalid code",
				skyerr.StringCause("CodeExpired"),
			)
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "invalid code",
				Code:    400,
				Info: map[string]interface{}{
					"cause": skyerr.StringCause("CodeExpired"),
				},
			})
		})
		Convey("error with causes", func() {
			NotAuthenticated := skyerr.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithCauses(
				"invalid password format",
				[]skyerr.Cause{
					skyerr.StringCause("TooShort"),
					skyerr.StringCause("TooSimple"),
				},
			)
			apiErr := skyerr.AsAPIError(err)
			So(apiErr, ShouldResemble, &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "invalid password format",
				Code:    400,
				Info: map[string]interface{}{
					"causes": []skyerr.Cause{
						skyerr.StringCause("TooShort"),
						skyerr.StringCause("TooSimple"),
					},
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
					"field": "email",
				},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Invalid","reason":"ValidationFailure","message":"failed to validate form payload","code":400,"info":{"field":"email"}}`)
		})
		Convey("error with causes", func() {
			apiErr := &skyerr.APIError{
				Kind:    skyerr.Kind{Name: skyerr.Invalid, Reason: "ValidationFailure"},
				Message: "invalid password format",
				Code:    400,
				Info: map[string]interface{}{
					"causes": []skyerr.Cause{
						skyerr.StringCause("TooShort"),
						skyerr.StringCause("TooSimple"),
					},
				},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Invalid","reason":"ValidationFailure","message":"invalid password format","code":400,"info":{"causes":[{"kind":"TooShort"},{"kind":"TooSimple"}]}}`)
		})
	})
}
