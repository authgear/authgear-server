package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type CreateIdentityRequest struct {
	Type model.IdentityType

	LoginID *CreateIdentityRequestLoginID
	OAuth   *CreateIdentityRequestOAuth
}

type CreateIdentityRequestOAuth struct {
	Alias string
	Spec  *identity.Spec
}

type CreateIdentityRequestLoginID struct {
	Spec *identity.Spec
}

func NewCreateOAuthIdentityRequest(alias string, spec *identity.Spec) *CreateIdentityRequest {
	return &CreateIdentityRequest{
		Type: model.IdentityTypeOAuth,
		OAuth: &CreateIdentityRequestOAuth{
			Alias: alias,
			Spec:  spec,
		},
	}
}

func NewCreateLoginIDIdentityRequest(spec *identity.Spec) *CreateIdentityRequest {
	return &CreateIdentityRequest{
		Type: model.IdentityTypeLoginID,
		LoginID: &CreateIdentityRequestLoginID{
			Spec: spec,
		},
	}
}
