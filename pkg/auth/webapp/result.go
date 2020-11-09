package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Result struct {
	redirectURI string
	cookies     []*http.Cookie
}

func (r *Result) WriteResponse(w http.ResponseWriter, req *http.Request) {
	for _, cookie := range r.cookies {
		httputil.UpdateCookie(w, cookie)
	}

	http.Redirect(w, req, r.redirectURI, http.StatusFound)
}

func (r *Result) IsInternalError() bool {
	return false
}
