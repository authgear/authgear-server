package skyerr

import (
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func ErrorFromRecoveringPanic(r interface{}) skyerr.Error {
	switch err := r.(type) {
	case skyerr.Error:
		return err
	case error:
		logrus.Errorf("%s", debug.Stack())
		return skyerr.NewErrorf(skyerr.UnexpectedError, "panic occurred while handling request: %v", err.Error())
	default:
		logrus.Warnf("router: unexpected type when recovering from panic: %v", err)
		return skyerr.NewErrorf(skyerr.UnexpectedError, "an panic occurred and the error is not known")
	}
}
