package httputil

import "net/http"

// FlashMessageTypeCookieDef is a HTTP session cookie.
var FlashMessageTypeCookieDef = CookieDef{
	Name:     "flash_message_type",
	Path:     "/",
	SameSite: http.SameSiteNoneMode,
}

type CookieFactoryInterface interface {
	ValueCookie(def *CookieDef, value string) *http.Cookie
	ClearCookie(def *CookieDef) *http.Cookie
}

type FlashMessage struct {
	CookieFactory CookieFactoryInterface
}

func (f *FlashMessage) Pop(r *http.Request, rw http.ResponseWriter) string {
	cookie, err := r.Cookie(FlashMessageTypeCookieDef.Name)
	if err != nil {
		return ""
	}

	messageType := cookie.Value

	clearCookie := f.CookieFactory.ClearCookie(&FlashMessageTypeCookieDef)
	UpdateCookie(rw, clearCookie)

	return messageType
}

func (f *FlashMessage) Flash(rw http.ResponseWriter, messageType string) {
	cookie := f.CookieFactory.ValueCookie(&FlashMessageTypeCookieDef, messageType)
	UpdateCookie(rw, cookie)
}
