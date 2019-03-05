package skyerr

import (
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func ErrorFromRecoveringPanic(r interface{}) Error {
	switch err := r.(type) {
	case Error:
		return err
	case error:
		logrus.Errorf("%s", debug.Stack())
		return NewErrorf(UnexpectedError, "panic occurred while handling request: %v", err.Error())
	default:
		logrus.Warnf("router: unexpected type when recovering from panic: %v", err)
		return NewErrorf(UnexpectedError, "an panic occurred and the error is not known")
	}
}
