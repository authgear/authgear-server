package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

// WeChatRedirectURICookieDef is a HTTP session cookie.
var WeChatRedirectURICookieDef = &httputil.CookieDef{
	NameSuffix: "wechat_redirect_uri",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// PlatformCookieDef is a HTTP session cookie.
var PlatformCookieDef = &httputil.CookieDef{
	NameSuffix: "platform",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// WeChatRedirectURIMiddleware validates x_wechat_redirect_uri and stores it in context.
// Ideally we should store x_wechat_redirect_uri in web app session.
// But we can link wechat in settings page so that is not possible at the moment.
type WeChatRedirectURIMiddleware struct {
	Cookies        CookieManager
	IdentityConfig *config.IdentityConfig
}

func (m *WeChatRedirectURIMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		weChatRedirectURI := q.Get("x_wechat_redirect_uri")
		isWechatEnabled := m.isWechatEnabled()
		if weChatRedirectURI != "" {
			// Validate x_wechat_redirect_uri
			valid := false
			for _, providerConfig := range m.IdentityConfig.OAuth.Providers {
				if providerConfig.Type != config.OAuthSSOProviderTypeWechat {
					continue
				}
				for _, allowed := range providerConfig.WeChatRedirectURIs {
					if weChatRedirectURI == allowed {
						valid = true
					}
				}
			}
			if isWechatEnabled && !valid {
				panic(apierrors.NewInvalid("wechat redirect URI is not allowed"))
			}
			if !isWechatEnabled {
				weChatRedirectURI = ""
			}
		}

		// Persist weChatRedirectURI.
		if weChatRedirectURI != "" {
			cookie := m.Cookies.ValueCookie(WeChatRedirectURICookieDef, weChatRedirectURI)
			httputil.UpdateCookie(w, cookie)
		}

		// Restore weChatRedirectURI from cookie
		if isWechatEnabled && weChatRedirectURI == "" {
			cookie, err := m.Cookies.GetCookie(r, WeChatRedirectURICookieDef)
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
			cookie := m.Cookies.ValueCookie(PlatformCookieDef, platform)
			httputil.UpdateCookie(w, cookie)
		}

		if platform == "" {
			cookie, err := m.Cookies.GetCookie(r, PlatformCookieDef)
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

func (m *WeChatRedirectURIMiddleware) isWechatEnabled() bool {
	for _, providerConfig := range m.IdentityConfig.OAuth.Providers {
		if providerConfig.Type != config.OAuthSSOProviderTypeWechat {
			continue
		}
		return true
	}
	return false
}
