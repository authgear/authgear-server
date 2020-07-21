package webapp

import (
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

type State struct {
	ID          string           `json:"id"`
	Error       *skyerr.APIError `json:"error"`
	RedirectURI string           `json:"redirect_uri,omitempty"`
	// ErrorRedirectURI is set by the previous step.
	// If graph is mutated successfully, it is reset to zero value.
	ErrorRedirectURI string `json:"error_redirect_uri,omitempty"`
	GraphInstanceID  string `json:"graph_instance_id,omitempty"`
}
