// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"net/http"
	"runtime/debug"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

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
	}[err.Code()]
	if !ok {
		if err.Code() < 10000 {
			log.Warnf("Error code %d (%v) does not have a default status code set. Assumed 500.", err.Code(), err.Code())
		}
		httpStatus = http.StatusInternalServerError
	}
	return httpStatus
}

func errorFromRecoveringPanic(r interface{}) skyerr.Error {
	switch err := r.(type) {
	case skyerr.Error:
		return err
	case error:
		log.Errorf("%s", debug.Stack())
		return skyerr.NewErrorf(skyerr.UnexpectedError, "panic occurred while handling request: %v", err.Error())
	default:
		log.Warnf("router: unexpected type when recovering from panic: %v", err)
		return skyerr.NewErrorf(skyerr.UnexpectedError, "an panic occurred and the error is not known")
	}
}
