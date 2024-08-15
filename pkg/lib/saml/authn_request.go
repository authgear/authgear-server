package saml

import (
	crewjamsaml "github.com/crewjam/saml"

	"github.com/authgear/authgear-server/pkg/lib/saml/binding"
)

type AuthnRequest struct {
	crewjamsaml.AuthnRequest
}

func (a *AuthnRequest) GetProtocolBinding() binding.SAMLBinding {
	return binding.SAMLBinding(a.AuthnRequest.ProtocolBinding)
}
