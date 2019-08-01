package skyerr

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func ErrorDefaultStatusCode(err Error) int {
	httpStatus, ok := map[ErrorCode]int{
		NotAuthenticated:        http.StatusUnauthorized,
		PermissionDenied:        http.StatusForbidden,
		AccessKeyNotAccepted:    http.StatusUnauthorized,
		AccessTokenNotAccepted:  http.StatusUnauthorized,
		InvalidCredentials:      http.StatusUnauthorized,
		InvalidSignature:        http.StatusUnauthorized,
		BadRequest:              http.StatusBadRequest,
		InvalidArgument:         http.StatusBadRequest,
		IncompatibleSchema:      http.StatusConflict,
		AtomicOperationFailure:  http.StatusConflict,
		PartialOperationFailure: http.StatusOK,
		Duplicated:              http.StatusConflict,
		ConstraintViolated:      http.StatusConflict,
		ResourceNotFound:        http.StatusNotFound,
		UndefinedOperation:      http.StatusNotFound,
		NotSupported:            http.StatusNotImplemented,
		NotImplemented:          http.StatusNotImplemented,
		PluginUnavailable:       http.StatusServiceUnavailable,
		PluginTimeout:           http.StatusGatewayTimeout,
		RecordQueryInvalid:      http.StatusBadRequest,
		ResponseTimeout:         http.StatusServiceUnavailable,
		DeniedArgument:          http.StatusForbidden,
		RecordQueryDenied:       http.StatusForbidden,
		NotConfigured:           http.StatusServiceUnavailable,
		PasswordPolicyViolated:  http.StatusBadRequest,
		UserDisabled:            http.StatusForbidden,
		VerificationRequired:    http.StatusForbidden,
		WebHookTimeOut:          http.StatusGatewayTimeout,
		WebHookFailed:           http.StatusBadGateway,
	}[err.Code()]
	if !ok {
		if err.Code() < 10000 {
			logrus.Warnf("Error code %d (%v) does not have a default status code set. Assumed 500.", err.Code(), err.Code())
		}
		httpStatus = http.StatusInternalServerError
	}
	return httpStatus
}
