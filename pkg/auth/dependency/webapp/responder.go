package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type ResponderInteractions interface {
	GetInteractionState(i *interaction.Interaction) (*interaction.State, error)
}

type Responder struct {
	ServerConfig  *config.ServerConfig
	StateProvider StateProvider
	Interactions  ResponderInteractions
}

type ErrorRedirect interface {
	RedirectURI() string
}

func (r *Responder) Respond(
	w http.ResponseWriter,
	req *http.Request,
	state *State,
	result *interactionflows.WebAppResult,
	err error,
) {
	onError := func() {
		if errorRedirect, ok := err.(ErrorRedirect); ok {
			http.Redirect(w, req, errorRedirect.RedirectURI(), http.StatusFound)
		} else {
			RedirectToCurrentPath(w, req)
		}
	}

	if err != nil {
		onError()
		return
	}

	for _, cookie := range result.Cookies {
		httputil.UpdateCookie(w, cookie)
	}

	if result.RedirectURI != "" {
		http.Redirect(w, req, result.RedirectURI, http.StatusFound)
		return
	}

	iState, err := r.Interactions.GetInteractionState(result.Interaction)
	if err != nil {
		onError()
		return
	}

	currentStep := iState.CurrentStep()
	switch currentStep.Step {
	case interaction.StepSetupPrimaryAuthenticator:
		switch currentStep.AvailableAuthenticators[0].Type {
		case authn.AuthenticatorTypeOOB:
			RedirectToPathWithX(w, req, "/oob_otp")
		case authn.AuthenticatorTypePassword:
			RedirectToPathWithX(w, req, "/create_password")
		default:
			panic("webapp: unexpected authenticator type")
		}
	case interaction.StepSetupSecondaryAuthenticator:
		panic("TODO: support StepSetupSecondaryAuthenticator")
	case interaction.StepAuthenticatePrimary:
		switch currentStep.AvailableAuthenticators[0].Type {
		case authn.AuthenticatorTypeOOB:
			RedirectToPathWithX(w, req, "/oob_otp")
		case authn.AuthenticatorTypePassword:
			RedirectToPathWithX(w, req, "/enter_password")
		default:
			panic("webapp: unexpected authenticator type")
		}
	case interaction.StepAuthenticateSecondary:
		panic("TODO: support StepAuthenticateSecondary")
	case interaction.StepCommit:
		r.StateProvider.DeleteState(state)
		RedirectToRedirectURI(w, req, r.ServerConfig.TrustProxy)
	default:
		panic("webapp: unexpected step")
	}
}
