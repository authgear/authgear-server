package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type Result interface {
	result() (*CompletionResult, error)
}

type CompletionResult struct {
	Client    config.OAuthClientConfiguration
	User      *model.User
	Principal *model.Identity
	Session   *model.Session

	SessionToken   string
	MFABearerToken string
}

func (r *CompletionResult) result() (*CompletionResult, error) {
	return r, nil
}

type InProgressResult struct {
	AuthnSessionToken string
	CurrentStep       SessionStep
}

func (r *InProgressResult) result() (*CompletionResult, error) {
	return nil, AuthenticationSessionRequired.NewWithInfo(
		"authentication session is required",
		skyerr.Details{"token": r.AuthnSessionToken, "step": r.CurrentStep},
	)
}
