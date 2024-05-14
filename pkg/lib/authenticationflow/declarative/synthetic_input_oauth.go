package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SyntheticInputOAuth struct {
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	Alias          string                                  `json:"alias,omitempty"`
	State          string                                  `json:"state,omitempty"`
	RedirectURI    string                                  `json:"redirect_uri,omitempty"`
	ResponseMode   string                                  `json:"response_mode,omitempty"`
	IdentitySpec   *identity.Spec                          `json:"identity_spec,omitempty"`
}

var _ authflow.Input = &SyntheticInputOAuth{}
var _ inputTakeIdentificationMethod = &SyntheticInputOAuth{}
var _ inputTakeOAuthAuthorizationRequest = &SyntheticInputOAuth{}
var _ syntheticInputOAuth = &SyntheticInputOAuth{}

func (*SyntheticInputOAuth) Input() {}

func (i *SyntheticInputOAuth) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *SyntheticInputOAuth) GetOAuthAlias() string {
	return i.Alias
}

func (i *SyntheticInputOAuth) GetOAuthState() string {
	return i.State
}

func (i *SyntheticInputOAuth) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

func (i *SyntheticInputOAuth) GetOAuthResponseMode() string {
	return i.ResponseMode
}

func (i *SyntheticInputOAuth) GetIdentitySpec() *identity.Spec {
	return i.IdentitySpec
}
