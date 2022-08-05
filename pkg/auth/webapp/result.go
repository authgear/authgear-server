package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Result struct {
	UILocales        string
	ColorScheme      string
	RedirectURI      string
	NavigationAction string
	Cookies          []*http.Cookie
	IsInteractionErr bool
}

func (r *Result) WriteResponse(w http.ResponseWriter, req *http.Request) {
	redirectURI, err := url.Parse(r.RedirectURI)
	if err != nil {
		panic(err)
	}

	q := redirectURI.Query()
	if r.UILocales != "" {
		q.Set("ui_locales", r.UILocales)
	}
	if r.ColorScheme != "" {
		q.Set("x_color_scheme", r.ColorScheme)
	}

	// Turbo now supports form submission natively.
	// We used to control turbo action "advance" or "replace" in the server side,
	// rather than in the client side.
	// The contract of selecting turbo action is by specifying
	// data-turbo-action in <button>, <a> or <form>.
	// Because we are responding a 30x response and
	// Turbo uses Fetch API which follows redirect transparently.
	// We CANNOT use header to convey data-turbo-action.
	// Instead, we write the intended turbo action as a query parameter.
	// In the client side, we have a Stimulus controller that listen for
	// "turbo:submit-end" event.
	// In the event, we can inspect the final response (i.e. all redirects followed)
	// to see if x_turbo_action is present.
	action := r.NavigationAction
	if action == "" {
		action = "advance"
	}
	q.Set("x_turbo_action", action)

	redirectURI.RawQuery = q.Encode()

	for _, cookie := range r.Cookies {
		httputil.UpdateCookie(w, cookie)
	}

	http.Redirect(w, req, redirectURI.String(), http.StatusFound)
}

func (r *Result) IsInternalError() bool {
	return false
}
