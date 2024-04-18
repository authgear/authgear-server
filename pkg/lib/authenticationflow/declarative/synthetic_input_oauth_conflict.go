package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// This input is for advancing the login flow with the conflicted existing identity
type SyntheticInputOAuthConflict struct {
	Identification config.AuthenticationFlowIdentification

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

func (*SyntheticInputOAuthConflict) Input() {}

var _ inputTakeIdentificationMethod = &SyntheticInputOAuthConflict{}
var _ inputTakeLoginID = &SyntheticInputOAuthConflict{}
