package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

//go:generate mockgen -source=dynamic_csp_middleware.go -destination=dynamic_csp_middleware_mock_test.go -package webapp

type AllowFrameAncestorsFromEnv bool

type AllowFrameAncestorsFromCustomUI bool

type DynamicCSPMiddleware struct {
	Cookies                         CookieManager
	OAuthConfig                     *config.OAuthConfig
	AllowedFrameAncestorsFromEnv    config.AllowedFrameAncestors
	AllowFrameAncestorsFromEnv      AllowFrameAncestorsFromEnv
	AllowFrameAncestorsFromCustomUI AllowFrameAncestorsFromCustomUI
}

func (m *DynamicCSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, r := httputil.CSPNoncePerSession(m.Cookies, w, r)

		var frameAncestors []string
		if m.AllowFrameAncestorsFromEnv {
			for _, frameAncestor := range m.AllowedFrameAncestorsFromEnv {
				frameAncestors = append(frameAncestors, frameAncestor)
			}
		}
		if m.AllowFrameAncestorsFromCustomUI {
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
			Nonce:          nonce,
			FrameAncestors: frameAncestors,
		})
		if err != nil {
			panic(err)
		}

		if len(frameAncestors) == 0 {
			w.Header().Set("X-Frame-Options", "DENY")
		}

		w.Header().Set("Content-Security-Policy", cspDirectives.String())
		next.ServeHTTP(w, r)
	})
}
