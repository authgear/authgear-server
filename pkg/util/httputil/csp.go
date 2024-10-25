package httputil

import (
	"context"
	"fmt"
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

type CSPKeywordSourceLevel1 string

var _ CSPSource = CSPKeywordSourceLevel1("")

const (
	CSPSourceNone         CSPKeywordSourceLevel1 = "'none'"
	CSPSourceSelf         CSPKeywordSourceLevel1 = "'self'"
	CSPSourceUnsafeInline CSPKeywordSourceLevel1 = "'unsafe-inline'"
	CSPSourceUnsafeEval   CSPKeywordSourceLevel1 = "'unsafe-eval'"
)

func (_ CSPKeywordSourceLevel1) CSPLevel() int {
	return 1
}

func (s CSPKeywordSourceLevel1) String() string {
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
	CSPDirectiveNameDefaultSrc           CSPDirectiveName = "default-src"
	CSPDirectiveNameScriptSrc            CSPDirectiveName = "script-src"
	CSPDirectiveNameWorkerSrc            CSPDirectiveName = "worker-src"
	CSPDirectiveNameObjectSrc            CSPDirectiveName = "object-src"
	CSPDirectiveNameBaseURI              CSPDirectiveName = "base-uri"
	CSPDirectiveNameBlockAllMixedContent CSPDirectiveName = "block-all-mixed-content"
	CSPDirectiveNameFrameAncestors       CSPDirectiveName = "frame-ancestors"
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
