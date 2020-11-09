package webapp

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Result struct {
	RedirectURI    string
	ReplaceCurrent bool
	Cookies        []*http.Cookie
}

func (r *Result) WriteResponse(w http.ResponseWriter, req *http.Request) {
	for _, cookie := range r.Cookies {
		httputil.UpdateCookie(w, cookie)
	}

	if req.Header.Get("X-Authgear-XHR") == "true" {
		type xhrResponse struct {
			RedirectURI string `json:"redirect_uri"`
			Replace     bool   `json:"replace"`
		}

		data, err := json.Marshal(xhrResponse{
			RedirectURI: r.RedirectURI,
			Replace:     r.ReplaceCurrent,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if _, err := w.Write(data); err != nil {
			panic(err)
		}
	} else {
		http.Redirect(w, req, r.RedirectURI, http.StatusFound)
	}
}

func (r *Result) IsInternalError() bool {
	return false
}
