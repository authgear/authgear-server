package apierrors_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/errorutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAPIError(t *testing.T) {
	Convey("AsAPIError", t, func() {
		Convey("simple error", func() {
			err := apierrors.NewInternalError("internal server error")
			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message: "internal server error",
				Code:    500,
				Info:    map[string]interface{}{},
			})
		})
		Convey("wrapped error", func() {
			var err error
			err = apierrors.NewInternalError("internal server error")
			err = fmt.Errorf("wrap this: %w", err)

			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message: "internal server error",
				Code:    500,
				Info:    map[string]interface{}{},
			})
		})
		Convey("common error", func() {
			NotAuthenticated := apierrors.Unauthorized.WithReason("NotAuthenticated")
			err := NotAuthenticated.New("authentication required")
			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Unauthorized, Reason: "NotAuthenticated"},
				Message: "authentication required",
				Code:    401,
				Info:    map[string]interface{}{},
			})
		})
		Convey("error with details", func() {
			NotAuthenticated := apierrors.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithDetails(
				"failed to validate form payload",
				apierrors.Details{
					"field":   apierrors.APIErrorDetail.Value("email"),
					"user_id": "user-id",
				},
			)
			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info: map[string]interface{}{
					"field": "email",
				},
			})
		})
		Convey("error with info", func() {
			NotAuthenticated := apierrors.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithInfo(
				"failed to validate form payload",
				apierrors.Details{"field": "email"},
			)
			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info: map[string]interface{}{
					"field": "email",
				},
			})
		})
		Convey("error with cause", func() {
			NotAuthenticated := apierrors.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithCause(
				"invalid code",
				apierrors.StringCause("CodeExpired"),
			)
			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "invalid code",
				Code:    400,
				Info: map[string]interface{}{
					"cause": apierrors.StringCause("CodeExpired"),
				},
			})
		})
		Convey("error with causes", func() {
			NotAuthenticated := apierrors.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithCauses(
				"invalid password format",
				[]apierrors.Cause{
					apierrors.StringCause("TooShort"),
					apierrors.StringCause("TooSimple"),
				},
			)
			apiErr := apierrors.AsAPIError(err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "invalid password format",
				Code:    400,
				Info: map[string]interface{}{
					"causes": []apierrors.Cause{
						apierrors.StringCause("TooShort"),
						apierrors.StringCause("TooSimple"),
					},
				},
			})
		})
		Convey("collect all details", func() {
			a := fmt.Errorf("a")
			b := errorutil.WithDetails(a, errorutil.Details{
				"b": apierrors.APIErrorDetail.Value("b"),
			})
			c := fmt.Errorf("c: %w", b)
			apiErr := apierrors.AsAPIError(c)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.InternalError, Reason: "UnexpectedError"},
				Message: "unexpected error occurred",
				Code:    apierrors.InternalError.HTTPStatus(),
				Info: map[string]interface{}{
					"b": "b",
				},
			})
		})
	})

	Convey("APIError", t, func() {
		Convey("simple error", func() {
			apiErr := &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message: "internal server error",
				Code:    500,
				Info:    map[string]interface{}{},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"InternalError","reason":"InternalError","message":"internal server error","code":500}`)
		})
		Convey("common error", func() {
			apiErr := &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Unauthorized, Reason: "NotAuthenticated"},
				Message: "authentication required",
				Code:    401,
				Info:    map[string]interface{}{},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Unauthorized","reason":"NotAuthenticated","message":"authentication required","code":401}`)
		})
		Convey("error with details", func() {
			apiErr := &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
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
			apiErr := &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "invalid password format",
				Code:    400,
				Info: map[string]interface{}{
					"causes": []apierrors.Cause{
						apierrors.StringCause("TooShort"),
						apierrors.StringCause("TooSimple"),
					},
				},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Invalid","reason":"ValidationFailure","message":"invalid password format","code":400,"info":{"causes":[{"kind":"TooShort"},{"kind":"TooSimple"}]}}`)
		})
	})
}
