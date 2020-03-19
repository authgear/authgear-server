package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type authnResolver interface {
	Resolve(
		client config.OAuthClientConfiguration,
		authnSessionToken string,
		stepPredicate func(authn.SessionStep) bool,
	) (*authn.Session, error)
}

type authnStepper interface {
	StepSession(
		client config.OAuthClientConfiguration,
		s authn.SessionContainer,
		mfaBearerToken string,
	) (authn.Result, error)

	WriteResult(rw http.ResponseWriter, result authn.Result)
}
