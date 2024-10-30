package portal

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func PortalCSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, r := httputil.CSPNoncePerRequest(r)
		data := map[string]interface{}{
			"CSPNonce": nonce,
		}
		r = r.WithContext(context.WithValue(r.Context(), httputil.FileServerIndexHTMLtemplateDataKey, data))
		cspDirectives := httputil.CSPDirectives{
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameScriptSrc,
				Value: httputil.CSPSources{
					// We used to include unsafe-eval here due to
					// https://github.com/facebook/regenerator/issues/336
					// and
					// https://github.com/facebook/regenerator/issues/450
					// But the two issues have been addressed since regenerator-runtime@0.13.8 (https://github.com/facebook/regenerator/commit/cc0cde9d90f975e5876df16c4b852c97f35da436)
					// If you run `rg regenerator-runtime` in ./portal you will see we are on regenerator-runtime@0.13.9
					// So we no longer need unsafe-eval anymore.

					// We intentionally do not support CSP1 browsers.
					// httputil.CSPSourceUnsafeInline,
					httputil.CSPSourceSelf,                // CSP1,CSP2
					httputil.CSPSchemeSourceHTTPS,         // CSP1,CSP2
					httputil.CSPNonceSource{Nonce: nonce}, // CSP2,CSP3
					httputil.CSPSourceStrictDynamic,       // CSP3
				},
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameObjectSrc,
				Value: httputil.CSPSources{
					httputil.CSPSourceNone, // CSP1,CSP2,CSP3
				},
			},
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameBaseURI,
				Value: httputil.CSPSources{
					httputil.CSPSourceNone, // CSP1,CSP2,CSP3
				},
			},
			// This must be kept in sync with httputil.XFrameOptionsDeny
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameFrameAncestors,
				Value: httputil.CSPSources{
					httputil.CSPSourceNone, // CSP2,CSP3
				},
			},
		}
		w.Header().Set("Content-Security-Policy", cspDirectives.String())
		next.ServeHTTP(w, r)
	})
}
