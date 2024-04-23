package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// This input is for advancing the login flow with the conflicted existing identity
type SyntheticInputOAuthConflict struct {
	Identification config.AuthenticationFlowIdentification

	// For identification=email/phone/username
	LoginID string

	// For identification=oauth
	Alias        string
	RedirectURI  string
	ResponseMode sso.ResponseMode
}

// GetLoginID implements inputTakeLoginID.
func (i *SyntheticInputOAuthConflict) GetLoginID() string {
	return i.LoginID
}

// GetIdentificationMethod implements inputTakeIdentificationMethod.
func (i *SyntheticInputOAuthConflict) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

// GetOAuthAlias implements inputTakeOAuthAuthorizationRequest.
func (i *SyntheticInputOAuthConflict) GetOAuthAlias() string {
	return i.Alias
}

// GetOAuthRedirectURI implements inputTakeOAuthAuthorizationRequest.
func (i *SyntheticInputOAuthConflict) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

// GetOAuthResponseMode implements inputTakeOAuthAuthorizationRequest.
func (i *SyntheticInputOAuthConflict) GetOAuthResponseMode() sso.ResponseMode {
	return i.ResponseMode
}

func (*SyntheticInputOAuthConflict) Input() {}

var _ inputTakeIdentificationMethod = &SyntheticInputOAuthConflict{}
var _ inputTakeLoginID = &SyntheticInputOAuthConflict{}
var _ inputTakeOAuthAuthorizationRequest = &SyntheticInputOAuthConflict{}
