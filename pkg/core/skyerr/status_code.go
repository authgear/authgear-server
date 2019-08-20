package skyerr

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func ErrorDefaultStatusCode(err Error) int {
	httpStatus, ok := map[ErrorCode]int{
		NotAuthenticated:            http.StatusUnauthorized,
		PermissionDenied:            http.StatusForbidden,
		AccessKeyNotAccepted:        http.StatusUnauthorized,
		AccessTokenNotAccepted:      http.StatusUnauthorized,
		InvalidCredentials:          http.StatusUnauthorized,
		BadRequest:                  http.StatusBadRequest,
		InvalidArgument:             http.StatusBadRequest,
		Duplicated:                  http.StatusConflict,
		ResourceNotFound:            http.StatusNotFound,
		UndefinedOperation:          http.StatusNotFound,
		PasswordPolicyViolated:      http.StatusBadRequest,
		UserDisabled:                http.StatusForbidden,
		VerificationRequired:        http.StatusForbidden,
		WebHookTimeOut:              http.StatusGatewayTimeout,
		WebHookFailed:               http.StatusBadGateway,
		CurrentIdentityBeingDeleted: http.StatusBadRequest,
	}[err.Code()]
	if !ok {
		if err.Code() < 10000 {
			logrus.Warnf("Error code %d (%v) does not have a default status code set. Assumed 500.", err.Code(), err.Code())
		}
		httpStatus = http.StatusInternalServerError
	}
	return httpStatus
}
