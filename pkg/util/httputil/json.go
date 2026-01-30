package httputil

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var JSONTooLarge = apierrors.RequestEntityTooLarge.WithReason("JSONTooLarge")

const BodyMaxSize = 1024 * 1024 * 10

type jsonOption struct {
	BodyMaxSize int64
}

type JSONOption func(option *jsonOption)

func makeDefaultOption() *jsonOption {
	return &jsonOption{
		BodyMaxSize: BodyMaxSize,
	}
}

func applyJSONOptions(options ...JSONOption) *jsonOption {
	option := makeDefaultOption()
	for _, o := range options {
		o(option)
	}
	return option
}

func WithBodyMaxSize(size int64) JSONOption {
	return func(option *jsonOption) {
		option.BodyMaxSize = size
	}
}

func IsJSONContentType(contentType string) bool {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	if mediaType != "application/json" {
		return false
	}
	// No params is good
	if len(params) == 0 {
		return true
	}
	// Contains unknown params
	if len(params) > 1 {
		return false
	}
	// The sole param must be charset=utf-8
	charset := params["charset"]
	return strings.ToLower(charset) == "utf-8"
}

func ParseJSONBody(r *http.Request, w http.ResponseWriter, parse func(io.Reader, interface{}) error, payload interface{}, options ...JSONOption) error {
	option := applyJSONOptions(options...)
	if !IsJSONContentType(r.Header.Get("Content-Type")) {
		return apierrors.NewBadRequest("request content type is invalid")
	}
	body := http.MaxBytesReader(w, r.Body, option.BodyMaxSize)
	defer body.Close()
	err := parse(body, payload)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return JSONTooLarge.NewWithInfo("request body too large", apierrors.Details{
				"limit": maxBytesError.Limit,
			})
		}

		return err
	}

	return nil
}

type BodyDefaulter interface {
	SetDefaults()
}

func BindJSONBody(r *http.Request, w http.ResponseWriter, v *validation.SchemaValidator, payload interface{}, options ...JSONOption) error {
	const errorMessage = "invalid request body"
	ctx := r.Context()
	return ParseJSONBody(r, w, func(reader io.Reader, value interface{}) error {
		err := v.ParseWithMessage(ctx, reader, errorMessage, value)
		if err != nil {
			return err
		}

		if value, ok := value.(BodyDefaulter); ok {
			value.SetDefaults()
		}
		return validation.ValidateValueWithMessage(ctx, value, errorMessage)
	}, payload, options...)
}

var JSONResponseWriterLogger = slogutil.NewLogger("json-response-writer")

func WriteJSONResponse(ctx context.Context, w http.ResponseWriter, resp *api.Response) {
	httpStatus := http.StatusOK
	apiError := apierrors.AsAPIErrorWithContext(ctx, resp.Error)
	if apiError != nil {
		httpStatus = apiError.Code
	}

	body, err := resp.EncodeToJSON(ctx)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if _, err := w.Write(body); err != nil {
		panic(err)
	}

	if apiError != nil && apiError.Code >= 500 && apiError.Code < 600 {
		logger := JSONResponseWriterLogger.GetLogger(ctx)
		logger.WithError(resp.Error).Error(ctx, "unexpected error occurred")
	}
}
