package mfa

import "errors"

var ErrDeviceTokenNotFound = errors.New("bearer token not found")

var ErrRecoveryCodeNotFound = errors.New("recovery code not found")
