package authenticator

import "errors"

var ErrAuthenticatorNotFound = errors.New("authenticator not found")

var ErrAuthenticatorAlreadyExists = errors.New("authenticator already exists")

var ErrInvalidCredentials = errors.New("invalid credentials")
