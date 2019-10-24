package presign

import (
	"net/http"
)

type Provider interface {
	Presign(r *http.Request)
	Verify(r *http.Request) error
}
