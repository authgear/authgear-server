package slogutil

import "errors"

// Internal errors we tracked as metric

var ErrOIDCDiscoveryInvalidStatusCode = errors.New("oidc discovery: invalid status code")
