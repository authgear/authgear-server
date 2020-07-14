package webapp

import (
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
)

type StateService interface {
	CreateState(s *interactionflows.State, redirectURI string) *interactionflows.State
	UpdateState(s *interactionflows.State, r *interactionflows.WebAppResult, inputError error)
	RestoreReadOnlyState(r *http.Request, optional bool) (state *interactionflows.State, err error)
	CloneState(r *http.Request) (state *interactionflows.State, err error)
}
