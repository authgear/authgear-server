package admin

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func AdminCSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, r := httputil.CSPNoncePerRequest(r)
		cspDirectives := httputil.CSPDirectives{
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameScriptSrc,
				Value: httputil.CSPSources{
					httputil.CSPSourceSelf,
					// CSP 1.
					httputil.CSPHostSource{
						Host: "unpkg.com",
					},
					// CSP 2.
					httputil.CSPNonceSource{
						Nonce: nonce,
					},
					// CSP 3.
					httputil.CSPSourceStrictDynamic,
				},
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameStyleSrc,
				Value: httputil.CSPSources{
					httputil.CSPSourceSelf,
					// CSP 1.
					httputil.CSPHostSource{
						Host: "unpkg.com",
					},
					// CSP 2.
					httputil.CSPNonceSource{
						Nonce: nonce,
					},
				},
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameObjectSrc,
				Value: httputil.CSPSources{
					httputil.CSPSourceNone,
				},
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameBaseURI,
				Value: httputil.CSPSources{
					httputil.CSPSourceNone,
				},
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameBlockAllMixedContent,
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameFrameAncestors,
				Value: httputil.CSPSources{
					httputil.CSPSourceNone,
				},
			},
		}
		w.Header().Set("Content-Security-Policy", cspDirectives.String())
		next.ServeHTTP(w, r)
	})
}
