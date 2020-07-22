package webapp

import (
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

type State struct {
	ID              string           `json:"id"`
	Error           *skyerr.APIError `json:"error"`
	RedirectURI     string           `json:"redirect_uri,omitempty"`
	KeepState       bool             `json:"keep_state,omitempty"`
	GraphInstanceID string           `json:"graph_instance_id,omitempty"`
}
