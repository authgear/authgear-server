package skyerr

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func ErrorDefaultStatusCode(err skyerr.Error) int {
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
		skyerr.PasswordPolicyViolated:  http.StatusBadRequest,
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
