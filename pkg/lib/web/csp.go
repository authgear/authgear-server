package web

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CSPDirectivesOptions struct {
	PublicOrigin string
	Nonce        string
	// FrameAncestors supports the redirect approach used by the custom UI.
	// The custom UI loads the redirect URI with an iframe.
	FrameAncestors []string
}

func CSPDirectives(opts CSPDirectivesOptions) (httputil.CSPDirectives, error) {
	u, err := url.Parse(opts.PublicOrigin)
	if err != nil {
		return nil, err
	}

	// We used to specify many host sources that we actually connect to.
	// But maintaining that list is troublesome.
	// So we now use the scheme source https: instead.
	//
	// Security mostly depends on the scripts that is allowed to run.
	// So getting script-src right is the most important.
	//
	// There are 3 CSP versions, namely CSP1, CSP2, and CSP3.
	// We want to make CSP3 browsers the most secure, while keeping
	// CSP1 browsers and CSP2 browsers still be able to function.
	//
	// The directives that will be effective in a CSP1 browser are
	// 'self' 'unsafe-inline' https:
	// That is CSP1 browser is vulnerable to XSS attack.
	//
	// The directives that will be effective in a CSP2 browser are
	// 'self' https: nonce- hash-
	// That is, 'unsafe-inline' will be ignored.
	//
	// The directives that will be effective in a CSP3 browser are
	// nonce- hash- 'strict-dynamic'
	// That is, 'unsafe-inline', host-sources, and scheme-sources will be ignored.

	scriptSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPSchemeSourceHTTPS,
		httputil.CSPNonceSource{
			Nonce: opts.Nonce,
		},
		httputil.CSPSourceStrictDynamic,
	}

	frameSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPSchemeSourceHTTPS,
	}

	fontSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPSchemeSourceHTTPS,
	}

	// style-src is also complicated.
	// Let me list out some notable cases here first.
	//
	// Turbo is known to write a stylesheet for ".turbo-progress-bar".
	// See https://github.com/hotwired/turbo/issues/809
	// To allow that to happen, we have two options.
	// 1. 'unsafe-eval'
	//   https://www.w3.org/TR/CSP2/#directive-style-src
	//   CSP2 says when 'unsafe-eval' is used, insertRule() can be used.
	// 2. Use hash-source, to explicitly allow the rule that Turbo is going to insert.
	//
	// We have some legit use cases of the style attribute that we cannot remove.
	// They are
	//   echo -n "position:absolute;width:0;height:0;" | openssl dgst -sha256 -binary | openssl enc -base64
	//   sha256-fOghyYcDMsLl/lf7piKeVgEljdV7IgqwGymlDo5oDhU=
	//
	//   echo -n "display:none;" | openssl dgst -sha256 -binary | openssl enc -base64
	//   sha256-0EZqoz+oBhx7gF4nvY2bSqoGyy4zLjNF+SDQXGp/ZrY=
	//
	//   echo -n "display:none;visibility:hidden;" | openssl dgst -sha256 -binary | openssl enc -base64
	//   sha256-ZLjZaRfcYelvFE+8S7ynGAe0XPN7SLX6dirEzdvD5Mk=
	//
	// To allow them, we have two options.
	// 1. 'unsafe-inline'. This works in CSP1, CSP2, and CSP3.
	//    https://www.w3.org/TR/CSP2/#directive-style-src
	//    CSP2 says 'unsafe-inline' allows the application of the style attribute.
	// 2. Use hash-source and 'unsafe-hashes'. But this only works in CSP3.
	//
	// We cannot use 1 and 2 at the same time because using a hash-source will make 'unsafe-inline' be ignored.
	// So using 2 implies we cannot use 1.
	// So to make things work in CSP1, CSP2, CSP3, we can only use 1.
	//
	// So the conclusion is that if we want to make things work in CSP1, CSP2, and CSP3, we need to have
	// style-src: unsafe-inline and unsafe-eval.

	styleSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPSourceUnsafeInline,
		httputil.CSPSourceUnsafeEval,
		httputil.CSPSchemeSourceHTTPS,
	}

	imgSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPSchemeSource{Scheme: "http"},
		httputil.CSPSchemeSourceHTTPS,
		// We use data URI to show QR image.
		// We can display external profile picture.
		httputil.CSPSchemeSource{Scheme: "data"},
	}

	// 'self' does not include websocket in Safari :(
	// https://github.com/w3c/webappsec-csp/issues/7
	connectSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPSchemeSourceHTTPS,
		httputil.CSPHostSource{
			Scheme: "ws",
			Host:   u.Host,
		},
		httputil.CSPHostSource{
			Scheme: "wss",
			Host:   u.Host,
		},
		// https://docs.sentry.io/platforms/javascript/install/cdn/#content-security-policy
		// The above doc says we need to specify `connect-src: *.sentry.io`,
		// But we already have `https:`, so that is no longer needed.
	}

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
			Name: httputil.CSPDirectiveNameDefaultSrc,
			Value: httputil.CSPSources{
				httputil.CSPSourceSelf,
			},
		},
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameScriptSrc,
			Value: scriptSrc,
		},
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameFrameSrc,
			Value: frameSrc,
		},
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameFontSrc,
			Value: fontSrc,
		},
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameStyleSrc,
			Value: styleSrc,
		},
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameImgSrc,
			Value: imgSrc,
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
			Name:  httputil.CSPDirectiveNameConnectSrc,
			Value: connectSrc,
		},
		httputil.CSPDirective{
			Name: httputil.CSPDirectiveNameBlockAllMixedContent,
		},
		httputil.CSPDirective{
			Name:  httputil.CSPDirectiveNameFrameAncestors,
			Value: frameAncestors,
		},
	}, nil
}
