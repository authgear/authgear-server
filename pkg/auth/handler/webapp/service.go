package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

// nolint:golint
type WebAppService interface {
	GetState(stateID string) (*webapp.State, error)
	GetIntent(webappIntent *webapp.Intent) (*webapp.State, *interaction.Graph, error)
	Get(stateID string) (*webapp.State, *interaction.Graph, error)
	PostIntent(webappIntent *webapp.Intent, inputer func() (interface{}, error)) (*webapp.Result, error)
	PostInput(stateID string, inputer func() (interface{}, error)) (*webapp.Result, error)
}

func StateID(r *http.Request) string {
	return r.Form.Get("x_sid")
}
