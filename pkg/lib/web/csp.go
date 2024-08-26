package web

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type CSPSource interface {
	CSPLevel() int
	String() string
}

type CSPKeywordSourceLevel3 string

var _ CSPSource = CSPKeywordSourceLevel3("")

const (
	CSPSourceStrictDynamic CSPKeywordSourceLevel3 = "'strict-dynamic'"
)

func (_ CSPKeywordSourceLevel3) CSPLevel() int {
	return 3
}

func (s CSPKeywordSourceLevel3) String() string {
	return string(s)
}

type CSPNonceSource struct {
	Nonce string
}

var _ CSPSource = CSPNonceSource{}

func (s CSPNonceSource) String() string {
	return fmt.Sprintf("'nonce-%v'", s.Nonce)
}

func (s CSPNonceSource) CSPLevel() int {
	return 2
}

type CSPSchemeSource struct {
	Scheme string
}

var _ CSPSource = CSPSchemeSource{}

func (s CSPSchemeSource) String() string {
	return fmt.Sprintf("%v:", s.Scheme)
}

func (s CSPSchemeSource) CSPLevel() int {
	return 1
}

type CSPHostSource struct {
	Scheme string
	Host   string
}

var _ CSPSource = CSPHostSource{}

func (s CSPHostSource) String() string {
	if s.Scheme != "" {
		return fmt.Sprintf("%v://%v", s.Scheme, s.Host)
	}
	return fmt.Sprintf("%v", s.Host)
}

func (s CSPHostSource) CSPLevel() int {
	return 1
}

type CSPKeywordSourceLevel1 string

var _ CSPSource = CSPKeywordSourceLevel1("")

const (
	CSPSourceNone         CSPKeywordSourceLevel1 = "'none'"
	CSPSourceSelf         CSPKeywordSourceLevel1 = "'self'"
	CSPSourceUnsafeInline CSPKeywordSourceLevel1 = "'unsafe-inline'"
)

func (_ CSPKeywordSourceLevel1) CSPLevel() int {
	return 1
}

func (s CSPKeywordSourceLevel1) String() string {
	return string(s)
}

type CSPSources []CSPSource

var _ sort.Interface = CSPSources{}

func (s CSPSources) Len() int {
	return len(s)
}

func (s CSPSources) Less(i, j int) bool {
	// Higher level source must appear before lower level source.
	return s[i].CSPLevel() > s[j].CSPLevel()
}

func (s CSPSources) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s CSPSources) String() string {
	var strs []string
	for _, source := range s {
		strs = append(strs, source.String())
	}
	return strings.Join(strs, " ")
}

type dynamicCSPContextKeyType struct{}

var dynamicCSPContextKey = dynamicCSPContextKeyType{}

type cspNonceContextValue struct {
	Nonce string
}

func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	v, ok := ctx.Value(dynamicCSPContextKey).(*cspNonceContextValue)
	if ok {
		v.Nonce = nonce
		return ctx
	}
	return context.WithValue(ctx, dynamicCSPContextKey, &cspNonceContextValue{Nonce: nonce})
}

func GetCSPNonce(ctx context.Context) string {
	v, ok := ctx.Value(dynamicCSPContextKey).(*cspNonceContextValue)
	if ok {
		return v.Nonce
	}
	return ""
}

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

var wwwgoogletagmanagercom = CSPHostSource{
	Host: "www.googletagmanager.com",
}

var euassetsiposthogcom = CSPHostSource{
	Host: "eu-assets.i.posthog.com",
}

var cdnjscloudflarecom = CSPHostSource{
	Host: "cdnjs.cloudflare.com",
}

var static2sharepointonlinecom = CSPHostSource{
	Host: "static2.sharepointonline.com",
}

var fontsgoogleapiscom = CSPHostSource{
	Host: "fonts.googleapis.com",
}

var fontsgstaticcom = CSPHostSource{
	Host: "fonts.gstatic.com",
}

var challengescloudflarecom = CSPHostSource{
	Host: "challenges.cloudflare.com",
}

var wwwgooglecom = CSPHostSource{
	Host: "www.google.com",
}

func CSPDirectives(opts CSPDirectivesOptions) ([]string, error) {
	u, err := url.Parse(opts.PublicOrigin)
	if err != nil {
		return nil, err
	}

	baseSrc := CSPSources{CSPSourceSelf}
	if opts.CDNHost != "" {
		baseSrc = append(baseSrc, CSPHostSource{
			Host: opts.CDNHost,
		})
	}

	// Unsafe-inline gets ignored if nonce is provided
	// https://w3c.github.io/webappsec-csp/#allow-all-inline
	var scriptSrc CSPSources
	if opts.AllowInlineScript {
		scriptSrc = CSPSources{
			CSPSourceUnsafeInline,
		}
	} else {
		scriptSrc = CSPSources{
			CSPSourceStrictDynamic,
			CSPNonceSource{
				Nonce: opts.Nonce,
			},
		}
	}
	scriptSrc = append(
		scriptSrc,
		euassetsiposthogcom,
		CSPHostSource{
			Scheme: "https",
			Host:   "browser.sentry-cdn.com",
		},
	)
	scriptSrc = append(scriptSrc, baseSrc...)
	sort.Sort(scriptSrc)

	frameSrc := CSPSources{
		wwwgoogletagmanagercom,
		challengescloudflarecom,
		wwwgooglecom,
		CSPSourceSelf,
	}

	fontSrc := CSPSources{
		cdnjscloudflarecom,
		static2sharepointonlinecom,
		fontsgoogleapiscom,
		fontsgstaticcom,
	}
	fontSrc = append(fontSrc, baseSrc...)
	sort.Sort(fontSrc)

	styleSrc := CSPSources{
		CSPSourceUnsafeInline,
		cdnjscloudflarecom,
		wwwgoogletagmanagercom,
		fontsgoogleapiscom,
	}
	styleSrc = append(styleSrc, baseSrc...)
	sort.Sort(styleSrc)

	imgSrc := CSPSources{
		CSPSchemeSource{
			Scheme: "http",
		},
		CSPSchemeSource{
			Scheme: "https",
		},
		// We use data URI to show QR image.
		// We can display external profile picture.
		CSPSchemeSource{
			Scheme: "data",
		},
	}
	imgSrc = append(imgSrc, baseSrc...)
	sort.Sort(imgSrc)

	// 'self' does not include websocket in Safari :(
	// https://github.com/w3c/webappsec-csp/issues/7
	connectSrc := CSPSources{
		CSPSourceSelf,
		CSPHostSource{
			Scheme: "https",
			Host:   "www.google-analytics.com",
		},
		CSPHostSource{
			Scheme: "ws",
			Host:   u.Host,
		},
		CSPHostSource{
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
		connectSrc = append(connectSrc, CSPHostSource{
			Host: sentryDSNHost,
		})
	}
	sort.Sort(connectSrc)

	var frameAncestors CSPSources
	if len(opts.FrameAncestors) > 0 {
		for _, host := range opts.FrameAncestors {
			frameAncestors = append(frameAncestors, CSPHostSource{
				Host: host,
			})
		}
	} else {
		frameAncestors = CSPSources{
			CSPSourceNone,
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
