package session

import (
	"net/http"
)

type Writer interface {
	WriteSession(rw http.ResponseWriter, accessToken *string, mfaBearerToken *string)
	ClearSession(rw http.ResponseWriter)
	ClearMFABearerToken(rw http.ResponseWriter)
}
