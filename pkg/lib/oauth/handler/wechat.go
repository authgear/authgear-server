package handler

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type weChatAuthzRequest interface {
	WeChatRedirectURI() string
}

func parseWeChatRedirectURI(client *config.OAuthClientConfig, r weChatAuthzRequest) (*url.URL, protocol.ErrorResponse) {
	allowedURIs := client.WeChatRedirectURIs
	redirectURIString := r.WeChatRedirectURI()
	// wechat redirect uri is optional
	if redirectURIString == "" {
		return nil, nil
	}

	redirectURI, err := url.Parse(redirectURIString)
	if err != nil {
		return nil, protocol.NewErrorResponse("invalid_request", "invalid wechat redirect URI")
	}

	allowed := false

	for _, u := range allowedURIs {
		if u == redirectURIString {
			allowed = true
			break
		}
	}

	if !allowed {
		return nil, protocol.NewErrorResponse("invalid_request", "wechat redirect URI is not allowed")
	}

	return redirectURI, nil
}
