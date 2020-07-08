package webapp

import (
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
)

type Responder interface {
	Respond(w http.ResponseWriter, r *http.Request, state *webapp.State, result *interactionflows.WebAppResult, err error)
}
