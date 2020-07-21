package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/httputil"
)

type Result struct {
	RedirectURI string
	Cookies     []*http.Cookie
}

func (r *Result) WriteResponse(w http.ResponseWriter, req *http.Request) {
	for _, cookie := range r.Cookies {
		httputil.UpdateCookie(w, cookie)
	}
	http.Redirect(w, req, r.RedirectURI, http.StatusFound)
}
