package oauth

import "errors"

var ErrAuthorizationNotFound = errors.New("oauth authorization not found")
var ErrGrantNotFound = errors.New("oauth grant not found")
