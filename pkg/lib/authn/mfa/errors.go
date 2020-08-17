package mfa

import "errors"

var ErrDeviceTokenNotFound = errors.New("bearer token not found")

var ErrRecoveryCodeNotFound = errors.New("recovery code not found")

var ErrRecoveryCodeConsumed = errors.New("recovery code consumed")
