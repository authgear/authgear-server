package e2eclient

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type FlowReference struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	URLQuery string `json:"url_query,omitempty"`
}

type FlowAction struct {
	Type           string                 `json:"type"`
	Identification string                 `json:"identification,omitempty"`
	Authentication string                 `json:"authentication,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
}

type FlowResponse struct {
	StateToken string      `json:"state_token"`
	Type       string      `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Action     *FlowAction `json:"action,omitempty"`
}

type HTTPResponse struct {
	Result *FlowResponse       `json:"result,omitempty"`
	Error  *apierrors.APIError `json:"error,omitempty"`
}

type SAMLBinding string

const (
	SAMLBindingHTTPRedirect SAMLBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
	SAMLBindingHTTPPost     SAMLBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
)
