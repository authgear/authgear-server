package nonce

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

var cookieDef = &httputil.CookieDef{
	Name:     "nonce",
	Path:     "/",
	SameSite: http.SameSiteNoneMode,
}

type Service struct {
	CookieFactory  *httputil.CookieFactory
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

func (s *Service) GenerateAndSet() string {
	n := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	cookie := s.CookieFactory.ValueCookie(cookieDef, n)
	httputil.UpdateCookie(s.ResponseWriter, cookie)
	return n
}

func (s *Service) GetAndClear() string {
	cookie, err := s.Request.Cookie(cookieDef.Name)
	if err != nil {
		return ""
	}
	n := cookie.Value
	cookie = s.CookieFactory.ClearCookie(cookieDef)
	httputil.UpdateCookie(s.ResponseWriter, cookie)
	return n
}
