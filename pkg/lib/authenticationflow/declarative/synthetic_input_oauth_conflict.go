package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// This input serves two purpose:
// 1. Pass oauth node using OAuthIdentitySpec, which contains the oauth identity that triggers the conflict
// 2. Advance the login flow with the conflicted existing identity
type SyntheticInputOAuthConflict struct {
	OAuthIdentitySpec *identity.Spec
	Identification    config.AuthenticationFlowIdentification

	// For identification=email/phone/username
	LoginID string
}

// GetLoginID implements inputTakeLoginID.
func (i *SyntheticInputOAuthConflict) GetLoginID() string {
	return i.LoginID
}

// GetIdentificationMethod implements inputTakeIdentificationMethod.
func (i *SyntheticInputOAuthConflict) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

// GetIdentitySpec implements syntheticInputOAuth.
func (i *SyntheticInputOAuthConflict) GetIdentitySpec() *identity.Spec {
	return i.OAuthIdentitySpec
}

func (*SyntheticInputOAuthConflict) Input() {}

var _ syntheticInputOAuth = &SyntheticInputOAuthConflict{}
var _ inputTakeIdentificationMethod = &SyntheticInputOAuthConflict{}
var _ inputTakeLoginID = &SyntheticInputOAuthConflict{}
