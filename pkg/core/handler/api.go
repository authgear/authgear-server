package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type APIHandler interface {
	DecodeRequest(request *http.Request) (RequestPayload, error)
	Handle(requestPayload interface{}, ctx context.AuthContext) (interface{}, error)
}

type APIResponse struct {
	Result interface{}  `json:"result,omitempty"`
	Err    skyerr.Error `json:"error,omitempty"`
}

func APIHandlerToHandler(apiHandler APIHandler) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, ctx context.AuthContext) {
		httpStatus := http.StatusOK
		response := APIResponse{}
		encoder := json.NewEncoder(rw)

		defer func() {
			if response.Err != nil {
				httpStatus = defaultStatusCode(response.Err)
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(httpStatus)
			encoder.Encode(response)
		}()

		payload, err := apiHandler.DecodeRequest(r)
		if err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		if err := payload.Validate(); err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		responsePayload, err := apiHandler.Handle(payload, ctx)
		if err == nil {
			response.Result = responsePayload
		} else {
			response.Err = skyerr.MakeError(err)
		}
	})
}

func defaultStatusCode(err skyerr.Error) int {
	httpStatus, ok := map[skyerr.ErrorCode]int{
		skyerr.NotAuthenticated:        http.StatusUnauthorized,
		skyerr.PermissionDenied:        http.StatusForbidden,
		skyerr.AccessKeyNotAccepted:    http.StatusUnauthorized,
		skyerr.AccessTokenNotAccepted:  http.StatusUnauthorized,
		skyerr.InvalidCredentials:      http.StatusUnauthorized,
		skyerr.InvalidSignature:        http.StatusUnauthorized,
		skyerr.BadRequest:              http.StatusBadRequest,
		skyerr.InvalidArgument:         http.StatusBadRequest,
		skyerr.IncompatibleSchema:      http.StatusConflict,
		skyerr.AtomicOperationFailure:  http.StatusConflict,
		skyerr.PartialOperationFailure: http.StatusOK,
		skyerr.Duplicated:              http.StatusConflict,
		skyerr.ConstraintViolated:      http.StatusConflict,
		skyerr.ResourceNotFound:        http.StatusNotFound,
		skyerr.UndefinedOperation:      http.StatusNotFound,
		skyerr.NotSupported:            http.StatusNotImplemented,
		skyerr.NotImplemented:          http.StatusNotImplemented,
		skyerr.PluginUnavailable:       http.StatusServiceUnavailable,
		skyerr.PluginTimeout:           http.StatusGatewayTimeout,
		skyerr.RecordQueryInvalid:      http.StatusBadRequest,
		skyerr.ResponseTimeout:         http.StatusServiceUnavailable,
		skyerr.DeniedArgument:          http.StatusForbidden,
		skyerr.RecordQueryDenied:       http.StatusForbidden,
		skyerr.NotConfigured:           http.StatusServiceUnavailable,
		skyerr.UserDisabled:            http.StatusForbidden,
		skyerr.VerificationRequired:    http.StatusForbidden,
	}[err.Code()]
	if !ok {
		if err.Code() < 10000 {
			logrus.Warnf("Error code %d (%v) does not have a default status code set. Assumed 500.", err.Code(), err.Code())
		}
		httpStatus = http.StatusInternalServerError
	}
	return httpStatus
}
