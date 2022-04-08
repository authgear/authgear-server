package httputil

import "net/http"

type TutorialCookieName string

const (
	SignupLoginTutorialCookieName TutorialCookieName = "signup_login_tutorial"
	SettingsTutorialCookieName    TutorialCookieName = "settings_tutorial"
)

var TutorialCookieNames = []TutorialCookieName{
	SignupLoginTutorialCookieName,
	SettingsTutorialCookieName,
}

type TutorialCookieManager interface {
	GetCookie(r *http.Request, def *CookieDef) (*http.Cookie, error)
	ValueCookie(def *CookieDef, value string) *http.Cookie
	ClearCookie(def *CookieDef) *http.Cookie
}

type TutorialCookie struct {
	Cookies FlashMessageCookieManager
}

func (t *TutorialCookie) Pop(r *http.Request, rw http.ResponseWriter, name TutorialCookieName) bool {
	cookieDef := makeCookieDef(name)

	cookie, err := t.Cookies.GetCookie(r, cookieDef)
	if err != nil {
		return false
	}
	v := cookie.Value

	clearCookie := t.Cookies.ClearCookie(cookieDef)
	UpdateCookie(rw, clearCookie)

	return v == "true"
}

func (t *TutorialCookie) SetAll(rw http.ResponseWriter) {
	for _, name := range TutorialCookieNames {
		cookieDef := makeCookieDef(name)
		cookie := t.Cookies.ValueCookie(cookieDef, "true")
		UpdateCookie(rw, cookie)
	}
}

func makeCookieDef(name TutorialCookieName) *CookieDef {
	return &CookieDef{
		NameSuffix: string(name),
		Path:       "/",
		SameSite:   http.SameSiteNoneMode,
	}
}
