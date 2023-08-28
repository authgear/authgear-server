package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

// CSPNonceCookieDef is a HTTP session cookie.
// The nonce has to be stable within a browsing session because
// Turbo uses XHR to load new pages.
// If nonce changes on every page load, the script in the new page
// cannot be run in the current page due to different nonce.
var CSPNonceCookieDef = &httputil.CookieDef{
	NameSuffix: "csp_nonce",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

type AllowInlineScript bool

type AllowFrameAncestors bool

type DynamicCSPMiddleware struct {
	Cookies             CookieManager
	HTTPOrigin          httputil.HTTPOrigin
	OAuthConfig         *config.OAuthConfig
	WebAppCDNHost       config.WebAppCDNHost
	AuthUISentryDSN     config.AuthUISentryDSN
	AllowInlineScript   AllowInlineScript
	AllowFrameAncestors AllowFrameAncestors
}

func (m *DynamicCSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nonce string
		cookie, err := m.Cookies.GetCookie(r, CSPNonceCookieDef)
		if err == nil {
			nonce = cookie.Value
		} else {
			nonce = rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
			cookie := m.Cookies.ValueCookie(CSPNonceCookieDef, nonce)
			httputil.UpdateCookie(w, cookie)
		}

		r = r.WithContext(web.WithCSPNonce(r.Context(), nonce))

		var frameAncestors []string
		if m.AllowFrameAncestors {
			for _, oauthClient := range m.OAuthConfig.Clients {
				if oauthClient.CustomUIURI != "" {
					u, err := url.Parse(oauthClient.CustomUIURI)
					if err != nil {
						panic(err)
					}
					frameAncestors = append(frameAncestors, urlutil.ExtractOrigin(u).String())
				}
			}
		}

		cspDirectives, err := web.CSPDirectives(web.CSPDirectivesOptions{
			PublicOrigin:      string(m.HTTPOrigin),
			Nonce:             nonce,
			CDNHost:           string(m.WebAppCDNHost),
			AuthUISentryDSN:   string(m.AuthUISentryDSN),
			AllowInlineScript: bool(m.AllowInlineScript),
			FrameAncestors:    frameAncestors,
		})
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Security-Policy", httputil.CSPJoin(cspDirectives))
		next.ServeHTTP(w, r)
	})
}
