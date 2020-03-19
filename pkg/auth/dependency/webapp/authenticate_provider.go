package webapp

import (
	"net/http"
)

type AuthenticateProvider interface {
	http.Handler
}
