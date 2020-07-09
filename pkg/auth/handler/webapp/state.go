package webapp

import (
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
)

type StateService interface {
	CreateState(r *http.Request, result *interactionflows.WebAppResult, inputError error) *interactionflows.State
	UpdateState(s *interactionflows.State, r *interactionflows.WebAppResult, inputError error)
	RestoreState(r *http.Request, optional bool) (state *interactionflows.State, err error)
}
