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
					// We intentionally do not support CSP1 browsers.
					// httputil.CSPSourceUnsafeInline,
					httputil.CSPSourceSelf,                // CSP1,CSP2
					httputil.CSPSchemeSourceHTTPS,         // CSP1,CSP2
					httputil.CSPNonceSource{Nonce: nonce}, // CSP2,CSP3
					httputil.CSPSourceStrictDynamic,       // CSP3
				},
			},
			httputil.CSPDirective{
				Name:  httputil.CSPDirectiveNameObjectSrc,
				Value: httputil.CSPSources{httputil.CSPSourceNone}, // CSP1,CSP2,CSP3
			},
			httputil.CSPDirective{
				Name:  httputil.CSPDirectiveNameBaseURI,
				Value: httputil.CSPSources{httputil.CSPSourceNone}, // CSP1,CSP2,CSP3
			},
			httputil.CSPDirective{
				Name:  httputil.CSPDirectiveNameFrameAncestors,
				Value: httputil.CSPSources{httputil.CSPSourceNone}, // CSP2,CSP3
			},
		}
		w.Header().Set("Content-Security-Policy", cspDirectives.String())
		next.ServeHTTP(w, r)
	})
}
