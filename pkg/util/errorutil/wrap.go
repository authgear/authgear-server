package errorutil

import (
	"fmt"
)

type errorWrap struct {
	inner error
	msg   string
}

func Wrap(err error, msg string) error {
	return &errorWrap{inner: err, msg: msg}
}

func Wrapf(err error, format string, args ...interface{}) error {
	return &errorWrap{inner: err, msg: fmt.Sprintf(format, args...)}
}

func (e *errorWrap) Error() string { return e.msg }
func (e *errorWrap) Unwrap() error { return e.inner }
