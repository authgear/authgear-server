package web

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CSPDirectivesOptions struct {
	PublicOrigin    string
	Nonce           string
	CDNHost         string
	AuthUISentryDSN string
	// FrameAncestors supports the redirect approach used by the custom UI.
	// The custom UI loads the redirect URI with an iframe.
	FrameAncestors    []string
	AllowInlineScript bool
}

var wwwgoogletagmanagercom = httputil.CSPHostSource{
	Host: "www.googletagmanager.com",
}

var euassetsiposthogcom = httputil.CSPHostSource{
	Host: "eu-assets.i.posthog.com",
}

var cdnjscloudflarecom = httputil.CSPHostSource{
	Host: "cdnjs.cloudflare.com",
}

var static2sharepointonlinecom = httputil.CSPHostSource{
	Host: "static2.sharepointonline.com",
}

var fontsgoogleapiscom = httputil.CSPHostSource{
	Host: "fonts.googleapis.com",
}

var fontsgstaticcom = httputil.CSPHostSource{
	Host: "fonts.gstatic.com",
}

var challengescloudflarecom = httputil.CSPHostSource{
	Host: "challenges.cloudflare.com",
}

var wwwgooglecom = httputil.CSPHostSource{
	Host: "www.google.com",
}

func CSPDirectives(opts CSPDirectivesOptions) ([]string, error) {
	u, err := url.Parse(opts.PublicOrigin)
	if err != nil {
		return nil, err
	}

	baseSrc := httputil.CSPSources{httputil.CSPSourceSelf}
	if opts.CDNHost != "" {
		baseSrc = append(baseSrc, httputil.CSPHostSource{
			Host: opts.CDNHost,
		})
	}

	// Unsafe-inline gets ignored if nonce is provided
	// https://w3c.github.io/webappsec-csp/#allow-all-inline
	var scriptSrc httputil.CSPSources
	if opts.AllowInlineScript {
		scriptSrc = httputil.CSPSources{
			httputil.CSPSourceUnsafeInline,
		}
	} else {
		scriptSrc = httputil.CSPSources{
			httputil.CSPSourceStrictDynamic,
			httputil.CSPNonceSource{
				Nonce: opts.Nonce,
			},
		}
	}
	scriptSrc = append(
		scriptSrc,
		wwwgoogletagmanagercom,
		euassetsiposthogcom,
		challengescloudflarecom,
		wwwgooglecom,
		httputil.CSPHostSource{
			Scheme: "https",
			Host:   "browser.sentry-cdn.com",
		},
	)
	scriptSrc = append(scriptSrc, baseSrc...)
	sort.Sort(scriptSrc)

	frameSrc := httputil.CSPSources{
		wwwgoogletagmanagercom,
		challengescloudflarecom,
		wwwgooglecom,
		httputil.CSPSourceSelf,
	}

	fontSrc := httputil.CSPSources{
		cdnjscloudflarecom,
		static2sharepointonlinecom,
		fontsgoogleapiscom,
		fontsgstaticcom,
	}
	fontSrc = append(fontSrc, baseSrc...)
	sort.Sort(fontSrc)

	styleSrc := httputil.CSPSources{
		httputil.CSPSourceUnsafeInline,
		cdnjscloudflarecom,
		wwwgoogletagmanagercom,
		fontsgoogleapiscom,
	}
	styleSrc = append(styleSrc, baseSrc...)
	sort.Sort(styleSrc)

	imgSrc := httputil.CSPSources{
		httputil.CSPSchemeSource{
			Scheme: "http",
		},
		httputil.CSPSchemeSource{
			Scheme: "https",
		},
		// We use data URI to show QR image.
		// We can display external profile picture.
		httputil.CSPSchemeSource{
			Scheme: "data",
		},
	}
	imgSrc = append(imgSrc, baseSrc...)
	sort.Sort(imgSrc)

	// 'self' does not include websocket in Safari :(
	// https://github.com/w3c/webappsec-csp/issues/7
	connectSrc := httputil.CSPSources{
		httputil.CSPSourceSelf,
		httputil.CSPHostSource{
			Scheme: "https",
			Host:   "www.google-analytics.com",
		},
		httputil.CSPHostSource{
			Scheme: "ws",
			Host:   u.Host,
		},
		httputil.CSPHostSource{
			Scheme: "wss",
			Host:   u.Host,
		},
	}
	// https://docs.sentry.io/platforms/javascript/install/cdn/#content-security-policy
	sentryDSNHost := ""
	if len(opts.AuthUISentryDSN) > 0 {
		u, err := url.Parse(opts.AuthUISentryDSN)
		if err != nil {
			return nil, fmt.Errorf("invalid AuthUISentryDSN %w", err)
		}
		sentryDSNHost = u.Host
	}
	if sentryDSNHost != "" {
		connectSrc = append(connectSrc, httputil.CSPHostSource{
			Host: sentryDSNHost,
		})
	}
	sort.Sort(connectSrc)

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

	return []string{
		"default-src 'self'",
		fmt.Sprintf("script-src %v", scriptSrc),
		fmt.Sprintf("frame-src %v", frameSrc),
		fmt.Sprintf("font-src %v", fontSrc),
		fmt.Sprintf("style-src %v", styleSrc),
		fmt.Sprintf("img-src %v", imgSrc),
		"object-src 'none'",
		"base-uri 'none'",
		fmt.Sprintf("connect-src %v", connectSrc),
		"block-all-mixed-content",
		fmt.Sprintf("frame-ancestors %v", frameAncestors),
	}, nil
}
