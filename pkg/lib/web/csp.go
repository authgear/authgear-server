package web

import (
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CSPDirectivesOptions struct {
	Nonce string
	// FrameAncestors supports the redirect approach used by the custom UI.
	// The custom UI loads the redirect URI with an iframe.
	FrameAncestors []string
}

func CSPDirectives(opts CSPDirectivesOptions) (httputil.CSPDirectives, error) {
	scriptSrc := httputil.CSPSources{
		// We intentionally do not support CSP1 browsers.
		// httputil.CSPSourceUnsafeInline,
		httputil.CSPSourceSelf,                     // CSP1,CSP2
		httputil.CSPSchemeSourceHTTPS,              // CSP1,CSP2
		httputil.CSPNonceSource{Nonce: opts.Nonce}, // CSP2,CSP3
		httputil.CSPSourceStrictDynamic,            // CSP3
	}

	// frame-src is no longer needed because we do not output default-src anymore.
	// Preivously, we output default-src so we have to specify frame-src to negate the effect of default-src.
	// frameSrc := httputil.CSPSources{
	// 	httputil.CSPSourceSelf,
	// 	httputil.CSPSchemeSourceHTTPS,
	// }

	// font-src is no longer needed because we do not output default-src anymore.
	// Preivously, we output default-src so we have to specify font-src to negate the effect of default-src.
	// fontSrc := httputil.CSPSources{
	// 	httputil.CSPSourceSelf,
	// 	httputil.CSPSchemeSourceHTTPS,
	// }

	// style-src is no longer needed because we do not output default-src anymore.
	// Preivously, we output default-src so we have to specify style-src to negate the effect of default-src.
	// styleSrc := httputil.CSPSources{
	// 	httputil.CSPSourceSelf,
	// 	httputil.CSPSourceUnsafeInline,
	// 	httputil.CSPSourceUnsafeEval,
	// 	httputil.CSPSchemeSourceHTTPS,
	// }

	// img-src is no longer needed because we do not output default-src anymore.
	// Preivously, we output default-src so we have to specify img-src to negate the effect of default-src.
	// imgSrc := httputil.CSPSources{
	// 	httputil.CSPSourceSelf,
	// 	httputil.CSPSchemeSource{Scheme: "http"},
	// 	httputil.CSPSchemeSourceHTTPS,
	// 	// We use data URI to show QR image.
	// 	// We can display external profile picture.
	// 	httputil.CSPSchemeSource{Scheme: "data"},
	// }

	// connect-src is no longer needed because we do not output default-src anymore.
	// Preivously, we output default-src so we have to specify connect-src to negate the effect of default-src.
	// connectSrc := httputil.CSPSources{
	// 	httputil.CSPSourceSelf,
	// 	httputil.CSPSchemeSourceHTTPS,
	// 	httputil.CSPHostSource{
	// 		Scheme: "ws",
	// 		Host:   u.Host,
	// 	},
	// 	httputil.CSPHostSource{
	// 		Scheme: "wss",
	// 		Host:   u.Host,
	// 	},
	// 	// https://docs.sentry.io/platforms/javascript/install/cdn/#content-security-policy
	// 	// The above doc says we need to specify `connect-src: *.sentry.io`,
	// 	// But we already have `https:`, so that is no longer needed.
	// }

	var frameAncestors httputil.CSPSources
	if len(opts.FrameAncestors) > 0 {
		for _, host := range opts.FrameAncestors {
			frameAncestors = append(frameAncestors, httputil.CSPHostSource{
				Host: host,
			})
		}
	} else {
		frameAncestors = httputil.CSPSources{
			httputil.CSPSourceNone,
		}
	}

	return httputil.CSPDirectives{
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameScriptSrc,
			Value: scriptSrc,
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
		// frame-ancestors is still needed to prevent from being iframed.
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameFrameAncestors,
			Value: frameAncestors,
		},
	}, nil
}
