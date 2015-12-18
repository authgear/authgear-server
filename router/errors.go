package router

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skyerr"
)

func defaultStatusCode(err skyerr.Error) int {
	switch err.Code() {
	case skyerr.NotAuthenticated:
	case skyerr.AccessKeyNotAccepted:
	case skyerr.AccessTokenNotAccepted:
	case skyerr.InvalidCredentials:
	case skyerr.InvalidSignature:
		return 401
	case skyerr.BadRequest:
	case skyerr.InvalidArgument:
	case skyerr.IncompatibleSchema:
		return 400
	case skyerr.Duplicated:
	case skyerr.ConstraintViolated:
		return 409
	case skyerr.ResourceNotFound:
	case skyerr.UndefinedOperation:
		return 404
	case skyerr.NotSupported:
	case skyerr.NotImplemented:
		return 501
	default:
		if err.Code() < 10000 {
			log.Warnf("Error code %d does not have a default status code set. Assumed 500.", err.Code())
		}
	}
	return 500
}
