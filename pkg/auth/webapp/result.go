package webapp

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type NavigationAction string

const (
	NavigationActionAdvance  NavigationAction = "advance"
	NavigationActionReplace  NavigationAction = "replace"
	NavigationActionRedirect NavigationAction = "redirect"
)

type Result struct {
	UILocales        string              `json:"ui_locales,omitempty"`
	ColorScheme      string              `json:"color_scheme,omitempty"`
	RedirectURI      string              `json:"redirect_uri,omitempty"`
	NavigationAction NavigationAction    `json:"navigation_action,omitempty"`
	Cookies          []*http.Cookie      `json:"cookies,omitempty"`
	IsInteractionErr bool                `json:"is_interaction_err,omitempty"`
	RemoveQueries    setutil.Set[string] `json:"remove_queries,omitempty"`
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

	// Navigate to self. Preserve query.
	if redirectURI.Host == "" {
		original := PreserveQuery(req.URL.Query())
		for key := range original {
			_, ignoreKey := r.RemoveQueries[key]
			// preserve the query only if it doesn't exist in the new query
			if q.Get(key) == "" && !ignoreKey {
				q.Set(key, original.Get(key))
			}
		}
	}

	// Propagate q_back_url
	// When user POST something to an endpoint, and that endpoint returned a redirect response,
	// from the user's perspective the user is reaching the redirected uri from the page he POST the request.
	if req.URL.Query().Get(QueryBackURL) != "" {
		q.Set(QueryBackURL, req.URL.Query().Get(QueryBackURL))
	}

	redirectURI.RawQuery = q.Encode()

	for _, cookie := range r.Cookies {
		httputil.UpdateCookie(w, cookie)
	}

	if req.Header.Get("X-Authgear-XHR") == "true" {
		type xhrResponse struct {
			RedirectURI string           `json:"redirect_uri"`
			Action      NavigationAction `json:"action"`
		}

		action := r.NavigationAction
		if action == "" {
			action = NavigationActionAdvance
		}
		data, err := json.Marshal(xhrResponse{
			RedirectURI: redirectURI.String(),
			Action:      action,
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
		http.Redirect(w, req, redirectURI.String(), http.StatusFound)
	}
}

func (r *Result) IsInternalError() bool {
	return false
}
