package httputil

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

// The spec of CSP1 is https://www.w3.org/TR/2012/CR-CSP-20121115/
// The spec of CSP2 is https://www.w3.org/TR/CSP2/
// The spec of CSP3 is https://www.w3.org/TR/CSP3/
//
//
// TL;DR: To implement strict CSP, follow this article.
// https://web.dev/articles/strict-csp#structure
//
//
// In CSP1, to protect against XSS, script-src and object-src is required.
// This is officially documented in https://www.w3.org/TR/2012/CR-CSP-20121115/#directives
// In CSP1, when default-src is present, then the user agent will enforce ALL directives defined in CSP1.
// This behavior is documented in https://www.w3.org/TR/2012/CR-CSP-20121115/#default-src
// Therefore, if default-src is present, that will bring a unwanted effect of restricting
// resources that are not essential to XSS protection.
// So when default-src is NOT present, it is unnecessary to specify
// - frame-src
// - font-src
// - style-src
// - img-src
// - connect-src
//
//
// If we specify style-src, we will get into the following troubles.
//
// Turbo is known to write a stylesheet for ".turbo-progress-bar".
// See https://github.com/hotwired/turbo/issues/809
// To allow that to happen, we have two options.
// 1. 'unsafe-eval'
//   https://www.w3.org/TR/CSP2/#directive-style-src
//   CSP2 says when 'unsafe-eval' is used, insertRule() can be used.
// 2. Use hash-source, to explicitly allow the rule that Turbo is going to insert.
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
//
//
// If we specify connect-src, we will get into the following troubles.
//
// 'self' in some previous versions of Safari is known to NOT include ws: and wss:
// See https://github.com/w3c/webappsec-csp/issues/7
//
//
// In CSP2, most changes are compatible with CSP1.
// The changes can be found in https://www.w3.org/TR/CSP2/#changes-from-level-1
// Our usage of CSP are not covered in the incompatible changes.
// In CSP2, frame-ancestors, nonce source, and hash source were introduced.
// frame-ancestors replaces X-Frame-Options.
// nonce source and hash source is mutually exclusive with 'unsafe-inline',
// meaning that if nonce source or hash source is present, then 'unsafe-inline' will be ignored
// by a CSP2 browser.
// Therefore, it is okay to have 'unsafe-inline' as long as either nonce source or hash source is present.
//
//
// In CSP3, the changes are compatible with CSP2.
// The changes can be found in https://www.w3.org/TR/CSP3/#changes-from-level-2
// In CSP3, 'strict-dynamic' was introduced.
// When 'strict-dynamic' is present, 'self', 'unsafe-inline', host source, and scheme source will be ignored, hash source and nonce source will be honored as usual.
// Therefore, it is okay to have 'self' and https: as long as 'strict-dynamic' is present.
//
// Given
//   script-src: 'unsafe-inline' 'self' https: nonce-NONCE 'strict-dynamic'
//
// In a CSP1 browser, it is interpreted as
//   script-src: 'unsafe-inline' 'self' https:
// Therefore, in a CSP1 browser, it is vulnerable to XSS attack.
// According to https://caniuse.com/contentsecuritypolicy2
// The following browsers support CSP2 (that is, they are not CSP1 browsers).
// - Chrome >= 40 (2015)
// - Edge >= 79 (2020)
// - Safari >= 10 (2016)
// - Firefox >= 46 (2016)
// - Android WebView >= 5.0 (2015)
// These browsers are covered the browserslist at ./authui/.browserslistrc.
//
// In a CSP2 browser, it is interpreted as
//   script-src: 'self' https: nonce-NONCE
// Therefore, in a CSP2 browser, it is vulnerable to XSS attack if the attack is able to
// inject a https: script into the HTML document.
// The attacker DOES NOT need to know the nonce to launch this attack due to the presence of https: scheme source.
// But the attacker CANNOT do this by injecting a inline script, because inline script is blocked, due to the absence of 'unsafe-inline'.
// Therefore, we can say, the attack is possible if the attacker can load a https: script in a nonced script.
// In other words, the attacker launches the attack in a trusted (nonced) script.
// This is not something that CSP can pretect.
//
// In a CSP3 browser, it is interpreted as
//   script-src: nonce-NONCE 'strict-dynamic'
// Therefore, in a CSP3 browser, it is not vulnerable to XSS attack.
// All scripts must be nonced in order to be executed by the browser.
// And the trust is propagated to the scripts further loaded by the nonced scripts.
// This behavior particularly useful in situation like Google Tag Manager,
// where you can define Custom HTML (which may contain a script tag).
// In CSP3, the custom HTML defined in GTM will be executed by the CSP3 browser.

// CSPNonceCookieDef is a HTTP session cookie.
// The nonce has to be stable within a browsing session because
// Turbo uses XHR to load new pages.
// If nonce changes on every page load, the script in the new page
// cannot be run in the current page due to different nonce.
var CSPNonceCookieDef = &CookieDef{
	NameSuffix: "csp_nonce",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

func makeCSPNonce() string {
	nonce := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return nonce
}

type CSPNoncePerSessionCookieManager interface {
	GetCookie(r *http.Request, def *CookieDef) (*http.Cookie, error)
	ValueCookie(def *CookieDef, value string) *http.Cookie
}

func CSPNoncePerSession(cookieManager CSPNoncePerSessionCookieManager, w http.ResponseWriter, r *http.Request) (nonce string, rWithNonce *http.Request) {
	cookie, err := cookieManager.GetCookie(r, CSPNonceCookieDef)
	if err == nil {
		nonce = cookie.Value
	} else {
		nonce = makeCSPNonce()
		cookie := cookieManager.ValueCookie(CSPNonceCookieDef, nonce)
		UpdateCookie(w, cookie)
	}

	rWithNonce = r.WithContext(WithCSPNonce(r.Context(), nonce))
	return
}

func CSPNoncePerRequest(r *http.Request) (nonce string, rWithNonce *http.Request) {
	nonce = makeCSPNonce()
	rWithNonce = r.WithContext(WithCSPNonce(r.Context(), nonce))
	return
}

type CSPSource interface {
	CSPLevel() int
	String() string
}

type CSPKeywordSourceLevel3 string

var _ CSPSource = CSPKeywordSourceLevel3("")

const (
	// 'unsafe-hashes' is not needed when we no longer specify style-src.
	// If you want it to allow inline event handler, you should migrate from inline event handler instead.
	// CSPSourceUnsafeHashes  CSPKeywordSourceLevel3 = "'unsafe-hashes'"
	CSPSourceStrictDynamic CSPKeywordSourceLevel3 = "'strict-dynamic'"
)

func (_ CSPKeywordSourceLevel3) CSPLevel() int {
	return 3
}

func (s CSPKeywordSourceLevel3) String() string {
	return string(s)
}

type CSPKeywordSourceLevel1 string

var _ CSPSource = CSPKeywordSourceLevel1("")

const (
	CSPSourceNone CSPKeywordSourceLevel1 = "'none'"
	CSPSourceSelf CSPKeywordSourceLevel1 = "'self'"
	// 'unsafe-inline' must be used with hash-source or nonce-source, and 'strict-dynamic'.
	// See the comment at the beginning of this file for details.
	// CSPSourceUnsafeInline CSPKeywordSourceLevel1 = "'unsafe-inline'"
	// 'unsafe-eval' is not needed when we no longer specify style-src.
	// CSPSourceUnsafeEval CSPKeywordSourceLevel1 = "'unsafe-eval'"
)

func (_ CSPKeywordSourceLevel1) CSPLevel() int {
	return 1
}

func (s CSPKeywordSourceLevel1) String() string {
	return string(s)
}

type CSPHashSource struct {
	Hash string
}

var _ CSPSource = CSPHashSource{}

func (s CSPHashSource) String() string {
	return fmt.Sprintf("'%v'", s.Hash)
}

func (s CSPHashSource) CSPLevel() int {
	return 2
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

var CSPSchemeSourceHTTPS = CSPSchemeSource{Scheme: "https"}

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

type CSPSources []CSPSource

var _ sort.Interface = CSPSources{}

func (s CSPSources) Len() int {
	return len(s)
}

func (s CSPSources) Less(i, j int) bool {
	// Previously, higher level source appears before lower level source.
	// But the spec https://www.w3.org/TR/CSP3/#strict-dynamic-usage says the opposite.
	// To deploy new directive in a compatible way, lower level source must appear before higher level source.
	return s[i].CSPLevel() < s[j].CSPLevel()
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

type CSPDirectiveName string

const (
	// default-src is not needed in implementing a strict CSP, and
	// it brings troubles.
	// See the comment at the beginning of this file for details.
	// CSPDirectiveNameDefaultSrc CSPDirectiveName = "default-src"

	// connect-src is not needed when there is no default-src.
	// CSPDirectiveNameConnectSrc CSPDirectiveName = "connect-src"
	// font-src is not needed when there is no default-src.
	// CSPDirectiveNameFontSrc    CSPDirectiveName = "font-src"
	// frame-src is not needed when there is no default-src.
	// CSPDirectiveNameFrameSrc   CSPDirectiveName = "frame-src"
	// img-src is not needed when there is no default-src.
	// CSPDirectiveNameImgSrc     CSPDirectiveName = "img-src"
	CSPDirectiveNameObjectSrc CSPDirectiveName = "object-src"
	CSPDirectiveNameScriptSrc CSPDirectiveName = "script-src"
	// style-src is not needed when there is no default-src.
	// CSPDirectiveNameStyleSrc   CSPDirectiveName = "style-src"
	// worker-src is not needed when there is no default-src.
	// CSPDirectiveNameWorkerSrc  CSPDirectiveName = "worker-src"

	CSPDirectiveNameBaseURI CSPDirectiveName = "base-uri"
	// CSPDirectiveNameBlockAllMixedContent is deprecated.
	// See https://www.w3.org/TR/mixed-content/#strict-checking
	// CSPDirectiveNameBlockAllMixedContent CSPDirectiveName = "block-all-mixed-content"
	CSPDirectiveNameFrameAncestors CSPDirectiveName = "frame-ancestors"
)

type CSPDirective struct {
	Name  CSPDirectiveName
	Value CSPSources
}

func (d CSPDirective) String() string {
	name := string(d.Name)
	if len(d.Value) <= 0 {
		return name
	}
	v := d.Value.String()
	return fmt.Sprintf("%v %v", name, v)
}

type CSPDirectives []CSPDirective

func (d CSPDirectives) String() string {
	var strs []string
	for _, directive := range d {
		strs = append(strs, directive.String())
	}
	return strings.Join(strs, "; ")
}
