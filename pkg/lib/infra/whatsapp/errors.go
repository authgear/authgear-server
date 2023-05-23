package whatsapp

import "errors"

var ErrUnexpectedStatus = errors.New("whatsapp: unexpected response status")
var ErrUnauthorized = errors.New("whatsapp: unauthorized")
var ErrUnexpectedLoginResponse = errors.New("whatsapp: unexpected login response body")
var ErrNoAvailableClient = errors.New("whatsapp: no available client")
