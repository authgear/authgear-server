package oauth

import "errors"

var ErrAuthorizationNotFound = errors.New("oauth authorization not found")
var ErrAuthorizationScopesNotGranted = errors.New("oauth authorization scopes not granted")
var ErrGrantNotFound = errors.New("oauth grant not found")
var ErrUnmatchedClient = errors.New("unmatched client id")
