package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type Result interface {
	result()
}

type CompletionResult struct {
	Client    config.OAuthClientConfiguration
	User      *model.User
	Principal *model.Identity
	Session   *model.Session

	SessionToken   string
	AccessToken    string
	RefreshToken   string
	MFABearerToken string
}

func (r *CompletionResult) result() {}

func (r *CompletionResult) UseCookie() bool {
	return r.Client == nil || r.Client.AuthAPIUseCookie()
}

type InProgressResult struct {
	AuthnSessionToken string
	CurrentStep       SessionStep
}

func (r *InProgressResult) result() {}

func (r *InProgressResult) ToAPIError() error {
	return AuthenticationSessionRequired.NewWithInfo(
		"authentication session is required",
		skyerr.Details{"token": r.AuthnSessionToken, "step": r.CurrentStep},
	)
}
