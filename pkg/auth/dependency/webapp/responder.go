package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type ResponderInteractions interface {
	GetInteractionState(i *interaction.Interaction) (*interaction.State, error)
}

type ResponderStates interface {
	DeleteState(*interactionflows.State)
}

type Responder struct {
	ServerConfig *config.ServerConfig
	States       ResponderStates
	Interactions ResponderInteractions
}

type ErrorRedirect interface {
	RedirectURI() string
}

func (r *Responder) Respond(
	w http.ResponseWriter,
	req *http.Request,
	state *interactionflows.State,
	result *interactionflows.WebAppResult,
	err error,
) {
	// result and err can be nil at the same time.

	// u is /req.URL.Path?x_sid=state.InstanceID
	u := state.RedirectURI(req.URL)

	onError := func() {
		if errorRedirect, ok := err.(ErrorRedirect); ok {
			http.Redirect(w, req, errorRedirect.RedirectURI(), http.StatusFound)
		} else {
			http.Redirect(w, req, u.String(), http.StatusFound)
		}
	}

	if err != nil {
		onError()
		return
	}

	if result != nil {
		for _, cookie := range result.Cookies {
			httputil.UpdateCookie(w, cookie)
		}
	}

	if result != nil && result.RedirectURI != "" {
		http.Redirect(w, req, result.RedirectURI, http.StatusFound)
		return
	}

	iState, err := r.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		onError()
		return
	}

	currentStep := iState.CurrentStep()
	switch currentStep.Step {
	case interaction.StepSetupPrimaryAuthenticator:
		switch currentStep.AvailableAuthenticators[0].Type {
		case authn.AuthenticatorTypeOOB:
			u.Path = "/oob_otp"
			http.Redirect(w, req, u.String(), http.StatusFound)
		case authn.AuthenticatorTypePassword:
			u.Path = "/create_password"
			http.Redirect(w, req, u.String(), http.StatusFound)
		default:
			panic("webapp: unexpected authenticator type")
		}
	case interaction.StepSetupSecondaryAuthenticator:
		panic("TODO: support StepSetupSecondaryAuthenticator")
	case interaction.StepAuthenticatePrimary:
		switch currentStep.AvailableAuthenticators[0].Type {
		case authn.AuthenticatorTypeOOB:
			u.Path = "/oob_otp"
			http.Redirect(w, req, u.String(), http.StatusFound)
		case authn.AuthenticatorTypePassword:
			u.Path = "/enter_password"
			http.Redirect(w, req, u.String(), http.StatusFound)
		default:
			panic("webapp: unexpected authenticator type")
		}
	case interaction.StepAuthenticateSecondary:
		panic("TODO: support StepAuthenticateSecondary")
	case interaction.StepCommit:
		r.States.DeleteState(state)

		redirectURI, _ := state.Extra[interactionflows.ExtraRedirectURI].(string)
		// Because we just deleted the state, we have to remove x_sid from redirectURI if it is present.
		u, err := url.Parse(redirectURI)
		if err == nil {
			q := u.Query()
			q.Del("x_sid")
			u.RawQuery = q.Encode()
			redirectURI = u.String()
		}

		http.Redirect(w, req, redirectURI, http.StatusFound)
	default:
		panic("webapp: unexpected step")
	}
}
