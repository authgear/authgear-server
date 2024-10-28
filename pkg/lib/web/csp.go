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
	FrameAncestors []string
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

func CSPDirectives(opts CSPDirectivesOptions) (httputil.CSPDirectives, error) {
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

	scriptSrc := httputil.CSPSources{
		httputil.CSPSourceStrictDynamic,
		httputil.CSPNonceSource{
			Nonce: opts.Nonce,
		},
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
		cdnjscloudflarecom,
		wwwgoogletagmanagercom,
		fontsgoogleapiscom,
		httputil.CSPHashSource{
			// https://github.com/hotwired/turbo/issues/809
			// Turbo is known to write a stylesheet for ".turbo-progress-bar".
			// Since we no longer allow unsafe-inline, we need another way to allow this.
			// The simplest way is to use hash source.
			// If turbo is upgraded, this hash is likely to change.
			// The way I obtained the hash is by trial-and-error.
			// I first omitted this hash source, then Chrome will complain about this,
			// and print the expected hash to console.
			Hash: "sha256-WAyOw4V+FqDc35lQPyRADLBWbuNK8ahvYEaQIYF1+Ps=",
		},
		httputil.CSPNonceSource{
			Nonce: opts.Nonce,
		},
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
