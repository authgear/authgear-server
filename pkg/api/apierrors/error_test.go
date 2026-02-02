package apierrors_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/errorutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAPIError(t *testing.T) {
	Convey("AsAPIError", t, func() {
		Convey("simple error", func() {
			err := apierrors.NewInternalError("internal server error")
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message:       "internal server error",
				Code:          500,
				TrackingID:    "",
				Info_ReadOnly: map[string]interface{}{},
			})
		})
		Convey("with tracking id", func() {
			err := apierrors.NewInternalError("internal server error")
			traceID := oteltrace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
			spanID := oteltrace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
			sc := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
				TraceID: traceID,
				SpanID:  spanID,
			})
			ctx := oteltrace.ContextWithSpanContext(context.Background(), sc)

			apiErr := apierrors.AsAPIErrorWithContext(ctx, err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message:       "internal server error",
				Code:          500,
				TrackingID:    "0102030405060708090a0b0c0d0e0f10-0102030405060708",
				Info_ReadOnly: map[string]interface{}{},
			})
		})
		Convey("with pre-existing tracking id", func() {
			err := &apierrors.APIError{
				Kind:       apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message:    "internal server error",
				Code:       500,
				TrackingID: "existing-id",
			}
			ctx := context.Background()

			apiErr := apierrors.AsAPIErrorWithContext(ctx, err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message:       "internal server error",
				Code:          500,
				TrackingID:    "existing-id",
				Info_ReadOnly: map[string]interface{}{},
			})
		})
		Convey("with tracking id attached by WithTrackingID", func() {
			traceID := oteltrace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
			spanID := oteltrace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
			sc := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
				TraceID: traceID,
				SpanID:  spanID,
			})
			ctx := oteltrace.ContextWithSpanContext(context.Background(), sc)

			err := errors.New("any error")
			errWithID := errorutil.WithTrackingID(ctx, err)

			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), errWithID)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.InternalError, Reason: "UnexpectedError"},
				Message:       "unexpected error occurred",
				Code:          500,
				TrackingID:    "0102030405060708090a0b0c0d0e0f10-0102030405060708",
				Info_ReadOnly: map[string]interface{}{},
			})
		})
		Convey("precedence of tracking id", func() {
			traceID1 := oteltrace.TraceID([16]byte{1})
			spanID1 := oteltrace.SpanID([8]byte{1})
			ctx1 := oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{TraceID: traceID1, SpanID: spanID1}))

			traceID2 := oteltrace.TraceID([16]byte{2})
			spanID2 := oteltrace.SpanID([8]byte{2})
			ctx2 := oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{TraceID: traceID2, SpanID: spanID2}))

			err := errors.New("any error")
			errWithID := errorutil.WithTrackingID(ctx1, err)

			// AsAPIErrorWithContext with ctx2 should still use ID from errWithID (ctx1)
			apiErr := apierrors.AsAPIErrorWithContext(ctx2, errWithID)
			So(apiErr.TrackingID, ShouldEqual, errorutil.FormatTrackingID(ctx1))
		})
		Convey("wrapped error", func() {
			var err error
			err = apierrors.NewInternalError("internal server error")
			err = fmt.Errorf("wrap this: %w", err)

			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message:       "internal server error",
				Code:          500,
				Info_ReadOnly: map[string]interface{}{},
			})
		})
		Convey("common error", func() {
			NotAuthenticated := apierrors.Unauthorized.WithReason("NotAuthenticated")
			err := NotAuthenticated.New("authentication required")
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.Unauthorized, Reason: "NotAuthenticated"},
				Message:       "authentication required",
				Code:          401,
				Info_ReadOnly: map[string]interface{}{},
			})
		})
		Convey("error with info", func() {
			NotAuthenticated := apierrors.Invalid.WithReason("ValidationFailure")
			err := NotAuthenticated.NewWithInfo(
				"failed to validate form payload",
				apierrors.Details{"field": "email"},
			)
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info_ReadOnly: map[string]interface{}{
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
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "invalid code",
				Code:    400,
				Info_ReadOnly: map[string]interface{}{
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
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "invalid password format",
				Code:    400,
				Info_ReadOnly: map[string]interface{}{
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
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), c)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.InternalError, Reason: "UnexpectedError"},
				Message: "unexpected error occurred",
				Code:    apierrors.InternalError.HTTPStatus(),
				Info_ReadOnly: map[string]interface{}{
					"b": "b",
				},
			})
		})

		Convey("recognize http.MaxBytesError", func() {
			err := &http.MaxBytesError{
				Limit: 1,
			}

			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.RequestEntityTooLarge, Reason: "RequestEntityTooLarge"},
				Message:       "http: request body too large",
				Code:          413,
				Info_ReadOnly: map[string]interface{}{},
			})
		})

		Convey("recognize JSON syntax error - case 1", func() {
			var unimportant interface{}
			err := json.Unmarshal([]byte(`{"a":}`), &unimportant)

			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.BadRequest, Reason: "InvalidJSON"},
				Message: "invalid character '}' looking for beginning of value",
				Code:    400,
				Info_ReadOnly: map[string]interface{}{
					"byte_offset": int64(6),
				},
			})
		})
		Convey("recognize JSON syntax error - case 2", func() {
			var unimportant interface{}
			err := json.Unmarshal([]byte(``), &unimportant)

			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			So(apiErr, ShouldResemble, &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.BadRequest, Reason: "InvalidJSON"},
				Message: "unexpected end of JSON input",
				Code:    400,
				Info_ReadOnly: map[string]interface{}{
					"byte_offset": int64(0),
				},
			})
		})
		Convey("it does not mutate the original error", func() {
			originalErr := apierrors.BadRequest.WithReason("test").New("testing error")

			newErr := errorutil.WithDetails(originalErr, errorutil.Details{"newkey": "test"})

			_ = apierrors.AsAPIErrorWithContext(context.Background(), newErr)

			// The original error info should not be modified
			So(originalErr.(*apierrors.APIError).Info_ReadOnly, ShouldResemble, make(apierrors.Details))

		})
	})

	Convey("IsAPIError", t, func() {
		Convey("simple error", func() {
			apiErr := apierrors.BadRequest.WithReason("Test").New("test")
			So(apierrors.IsAPIError(apiErr), ShouldEqual, true)
		})

		Convey("joined error with api error in the front", func() {
			apiErr := apierrors.BadRequest.WithReason("Test").New("test")
			joinedError := errors.Join(apiErr, errors.New("test"))
			So(apierrors.IsAPIError(joinedError), ShouldEqual, true)
		})

		Convey("joined error with api error in the end", func() {
			apiErr := apierrors.BadRequest.WithReason("Test").New("test")
			joinedError := errors.Join(errors.New("test"), apiErr)
			So(apierrors.IsAPIError(joinedError), ShouldEqual, true)
		})
	})

	Convey("APIError", t, func() {
		Convey("simple error", func() {
			apiErr := &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.InternalError, Reason: string(apierrors.InternalError)},
				Message:       "internal server error",
				Code:          500,
				Info_ReadOnly: map[string]interface{}{},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"InternalError","reason":"InternalError","message":"internal server error","code":500}`)
		})
		Convey("common error", func() {
			apiErr := &apierrors.APIError{
				Kind:          apierrors.Kind{Name: apierrors.Unauthorized, Reason: "NotAuthenticated"},
				Message:       "authentication required",
				Code:          401,
				Info_ReadOnly: map[string]interface{}{},
			}
			json, _ := json.Marshal(apiErr)
			So(string(json), ShouldEqual, `{"name":"Unauthorized","reason":"NotAuthenticated","message":"authentication required","code":401}`)
		})
		Convey("error with details", func() {
			apiErr := &apierrors.APIError{
				Kind:    apierrors.Kind{Name: apierrors.Invalid, Reason: "ValidationFailure"},
				Message: "failed to validate form payload",
				Code:    400,
				Info_ReadOnly: map[string]interface{}{
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
				Info_ReadOnly: map[string]interface{}{
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
