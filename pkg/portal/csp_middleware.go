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
					httputil.CSPSourceSelf,
					// We used to include unsafe-eval here due to
					// https://github.com/facebook/regenerator/issues/336
					// and
					// https://github.com/facebook/regenerator/issues/450
					// But the two issues have been addressed since regenerator-runtime@0.13.8 (https://github.com/facebook/regenerator/commit/cc0cde9d90f975e5876df16c4b852c97f35da436)
					// If you run `rg regenerator-runtime` in ./portal you will see we are on regenerator-runtime@0.13.9
					// So we no longer need unsafe-eval anymore.
					httputil.CSPHostSource{
						Host: "cdn.jsdelivr.net",
					},
					httputil.CSPHostSource{
						Host: "unpkg.com",
					},
					httputil.CSPHostSource{
						Host: "www.googletagmanager.com",
					},
					httputil.CSPHostSource{
						Host: "cdn.mxpnl.com",
					},
					httputil.CSPHostSource{
						Host: "eu.posthog.com",
					},
					httputil.CSPHostSource{
						Host: "eu-assets.i.posthog.com",
					},
					httputil.CSPHostSource{
						Host: "cmp.osano.com",
					},
					httputil.CSPNonceSource{
						Nonce: nonce,
					},
					httputil.CSPSourceStrictDynamic,
				},
			},
			// monaco editor create worker with blob:
			httputil.CSPDirective{
				Name: httputil.CSPDirectiveNameWorkerSrc,
				Value: httputil.CSPSources{
					httputil.CSPSourceSelf,
					httputil.CSPHostSource{
						Host: "cdn.jsdelivr.net",
					},
					httputil.CSPSchemeSource{
						Scheme: "blob",
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
			// This must be kept in sync with httputil.XFrameOptionsDeny
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
