package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

// WeChatRedirectURICookieDef is a HTTP session cookie.
var WeChatRedirectURICookieDef = httputil.CookieDef{
	Name:     "wechat_redirect_uri",
	Path:     "/",
	SameSite: http.SameSiteNoneMode,
}

// PlatformCookieDef is a HTTP session cookie.
var PlatformCookieDef = httputil.CookieDef{
	Name:     "platform",
	Path:     "/",
	SameSite: http.SameSiteNoneMode,
}

type WeChatRedirectURIMiddleware struct {
	CookieFactory CookieFactory
}

func (m *WeChatRedirectURIMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		weChatRedirectURI := q.Get("x_wechat_redirect_uri")

		// Persist weChatRedirectURI.
		if weChatRedirectURI != "" {
			cookie := m.CookieFactory.ValueCookie(&WeChatRedirectURICookieDef, weChatRedirectURI)
			httputil.UpdateCookie(w, cookie)
		}

		// Restore weChatRedirectURI from cookie
		if weChatRedirectURI == "" {
			cookie, err := r.Cookie(WeChatRedirectURICookieDef.Name)
			if err == nil {
				weChatRedirectURI = cookie.Value
			}
		}

		// Restore weChatRedirectURI into the request context.
		if weChatRedirectURI != "" {
			ctx := wechat.WithWeChatRedirectURI(r.Context(), weChatRedirectURI)
			r = r.WithContext(ctx)
		}

		// Repeat the steps for platform
		platform := q.Get("x_platform")
		if platform != "" {
			cookie := m.CookieFactory.ValueCookie(&PlatformCookieDef, platform)
			httputil.UpdateCookie(w, cookie)
		}

		if platform == "" {
			cookie, err := r.Cookie(PlatformCookieDef.Name)
			if err == nil {
				platform = cookie.Value
			}
		}

		if platform != "" {
			ctx := wechat.WithPlatform(r.Context(), platform)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
