package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AccountRecoveryIdentificationOption struct {
	Identification config.AuthenticationFlowRequestAccountRecoveryIdentification `json:"identification"`
}

type AccountRecoveryChannel string

const (
	AccountRecoveryChannelEmail AccountRecoveryChannel = "email"
	AccountRecoveryChannelSMS   AccountRecoveryChannel = "sms"
)

type AccountRecoveryDestinationOption struct {
	ID                string                 `json:"id"`
	MaskedDisplayName string                 `json:"masked_display_name"`
	Channel           AccountRecoveryChannel `json:"channel"`
}

type AccountRecoveryDestinationOptionInternal struct {
	AccountRecoveryDestinationOption
	TargetLoginID string `json:"target_login_id"`
}

type AccountRecoveryIdentity struct {
	Identification config.AuthenticationFlowRequestAccountRecoveryIdentification
	IdentitySpec   *identity.Spec
	MaybeIdentity  *identity.Info
}
