package webapp

import (
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
)

type Responder interface {
	Respond(w http.ResponseWriter, r *http.Request, state *interactionflows.State, result *interactionflows.WebAppResult, err error)
}
