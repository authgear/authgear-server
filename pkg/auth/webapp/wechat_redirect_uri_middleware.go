package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

// ClientCookieDef is a HTTP session cookie.
var ClientCookieDef = httputil.CookieDef{
	Name:     "client_id",
	Path:     "/",
	SameSite: http.SameSiteNoneMode,
}

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
	Config        *config.OAuthConfig
}

func (m *WeChatRedirectURIMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		// We need the client id information check if parse wechat_redirect_uri
		// is allowed
		clientID := q.Get("client_id")
		// store client id into cookie.
		if clientID != "" {
			cookie := m.CookieFactory.ValueCookie(&ClientCookieDef, clientID)
			httputil.UpdateCookie(w, cookie)
		}

		// Restore client id per checking
		if clientID == "" {
			cookie, err := r.Cookie(ClientCookieDef.Name)
			if err == nil {
				clientID = cookie.Value
			}
		}

		client := resolveClient(m.Config, clientID)

		// We need to check and ensure wechat_redirect_uri is valid for
		// the given oauth client
		// if weChatRedirectURI parameter exists in the request
		// return error if weChatRedirectURI is invalid

		// if weChatRedirectURI is restored from cookie
		// we only use the uri if it passes the checking
		// if it failed the checking, just skip it and ingore the error
		// we don't want the invalid uri in cookie value block the flow
		// permanently, we just don't use the invalid value
		weChatRedirectURI := q.Get("x_wechat_redirect_uri")
		if weChatRedirectURI != "" {
			err := parseWeChatRedirectURI(client, weChatRedirectURI)
			if err != nil {
				http.Error(w, fmt.Sprintf("oauth: %s", err.Error()), http.StatusBadRequest)
				return
			}
		} else {
			cookie, err := r.Cookie(WeChatRedirectURICookieDef.Name)
			if err == nil {
				cookieValue := cookie.Value
				err = parseWeChatRedirectURI(client, cookieValue)
				if err == nil {
					// valid weChatRedirectURI in cookie
					// we can use it
					weChatRedirectURI = cookieValue
				}
			}
		}

		if weChatRedirectURI != "" {
			// Persist weChatRedirectURI.
			cookie := m.CookieFactory.ValueCookie(&WeChatRedirectURICookieDef, weChatRedirectURI)
			httputil.UpdateCookie(w, cookie)

			// Restore weChatRedirectURI the request context.
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

func parseWeChatRedirectURI(client *config.OAuthClientConfig, weChatRedirectURI string) error {
	if client == nil {
		return errors.New("invalid client id for wechat redirect URI")
	}

	allowedURIs := client.WeChatRedirectURIs
	// wechat redirect uri is optional
	if weChatRedirectURI == "" {
		return nil
	}

	_, err := url.Parse(weChatRedirectURI)
	if err != nil {
		return errors.New("invalid wechat redirect URI")
	}

	allowed := false

	for _, u := range allowedURIs {
		if u == weChatRedirectURI {
			allowed = true
			break
		}
	}

	if !allowed {
		return errors.New("wechat redirect URI is not allowed")
	}

	return nil
}

func resolveClient(config *config.OAuthConfig, clientID string) *config.OAuthClientConfig {
	if clientID == "" {
		return nil
	}
	if client, ok := config.GetClient(clientID); ok {
		return client
	}
	return nil
}
