package webappoauth

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type WebappOAuthState struct {
	AppID            string                  `json:"app_id"`
	UIImplementation config.UIImplementation `json:"ui_implementation"`
	WebSessionID     string                  `json:"web_session_id"`

	// authflow, authflowv2 specific fields
	XStep            string `json:"x_step"`
	ErrorRedirectURI string `json:"error_redirect_uri"`

	// account management specific fields
	AccountManagementToken string `json:"account_management_token,omitempty"`
}
