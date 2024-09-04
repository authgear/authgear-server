package nonce

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

var cookieDef = &httputil.CookieDef{
	NameSuffix:    "nonce",
	Path:          "/",
	SameSite:      http.SameSiteNoneMode,
	IsNonHostOnly: false,
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type Service struct {
	Cookies        CookieManager
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

func (s *Service) GenerateAndSet() string {
	n := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	cookie := s.Cookies.ValueCookie(cookieDef, n)
	httputil.UpdateCookie(s.ResponseWriter, cookie)
	return n
}

func (s *Service) GetAndClear() string {
	cookie, err := s.Cookies.GetCookie(s.Request, cookieDef)
	if err != nil {
		return ""
	}
	n := cookie.Value
	cookie = s.Cookies.ClearCookie(cookieDef)
	httputil.UpdateCookie(s.ResponseWriter, cookie)
	return n
}
