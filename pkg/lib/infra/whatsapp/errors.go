package whatsapp

import "errors"

var ErrUnexpectedStatus = errors.New("whatsapp: unexpected response status")
var ErrUnexpectedLoginResponse = errors.New("whatsapp: unexpected login response body")
