package httputil

import (
	"net/http"
)

// FlashMessageTypeCookieDef is a HTTP session cookie.
var FlashMessageTypeCookieDef = &CookieDef{
	NameSuffix:    "flash_message_type",
	Path:          "/",
	SameSite:      http.SameSiteNoneMode,
	IsNonHostOnly: false,
}

type FlashMessageCookieManager interface {
	GetCookie(r *http.Request, def *CookieDef) (*http.Cookie, error)
	ValueCookie(def *CookieDef, value string) *http.Cookie
	ClearCookie(def *CookieDef) *http.Cookie
}

type FlashMessage struct {
	Cookies FlashMessageCookieManager
}

func (f *FlashMessage) Pop(r *http.Request, rw http.ResponseWriter) string {
	cookie, err := f.Cookies.GetCookie(r, FlashMessageTypeCookieDef)
	if err != nil {
		return ""
	}

	messageType := cookie.Value

	clearCookie := f.Cookies.ClearCookie(FlashMessageTypeCookieDef)
	UpdateCookie(rw, clearCookie)

	return messageType
}

func (f *FlashMessage) Flash(rw http.ResponseWriter, messageType string) {
	cookie := f.Cookies.ValueCookie(FlashMessageTypeCookieDef, messageType)
	UpdateCookie(rw, cookie)
}
