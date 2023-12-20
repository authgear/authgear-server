package whatsapp

import "errors"

var ErrUnauthorized = errors.New("whatsapp: unauthorized")
var ErrInvalidUser = errors.New("whatsapp: invalid user")
var ErrBadRequest = errors.New("whatsapp: bad request")
var ErrUnexpectedLoginResponse = errors.New("whatsapp: unexpected login response body")
var ErrNoAvailableClient = errors.New("whatsapp: no available client")
