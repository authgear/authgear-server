package webapp

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

type State struct {
	ID              string                 `json:"id"`
	PrevID          string                 `json:"prev_id"`
	Error           *skyerr.APIError       `json:"error"`
	RedirectURI     string                 `json:"redirect_uri,omitempty"`
	KeepState       bool                   `json:"keep_state,omitempty"`
	GraphInstanceID string                 `json:"graph_instance_id,omitempty"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
	UserAgentToken  string                 `json:"user_agent_token"`
	UILocales       string                 `json:"ui_locales,omitempty"`
}

func (s *State) SetID(id string) {
	s.PrevID = s.ID
	s.ID = id
}

func AttachStateID(id string, input *url.URL) *url.URL {
	u := *input

	q := u.Query()
	q.Set("x_sid", id)

	u.Scheme = ""
	u.Opaque = ""
	u.Host = ""
	u.User = nil
	u.RawQuery = q.Encode()

	return &u
}
