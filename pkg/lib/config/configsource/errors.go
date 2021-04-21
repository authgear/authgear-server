package configsource

import "errors"

var ErrAppNotFound = errors.New("app not found")
var ErrDuplicatedAppID = errors.New("duplicated app ID")
