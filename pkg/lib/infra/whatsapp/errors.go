package whatsapp

import "errors"

var ErrUnauthorized = errors.New("whatsapp: unauthorized")
var ErrUnexpectedLoginResponse = errors.New("whatsapp: unexpected login response body")
var ErrNoAvailableClient = errors.New("whatsapp: no available client")
