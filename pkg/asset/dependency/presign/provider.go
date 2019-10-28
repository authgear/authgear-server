package presign

import (
	"net/http"
	"time"
)

type Provider interface {
	Presign(r *http.Request, expires time.Duration)
	Verify(r *http.Request) error
}
