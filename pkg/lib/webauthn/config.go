package webauthn

import (
	"github.com/duo-labs/webauthn/protocol"
)

type Config struct {
	RPID                        string
	RPOrigin                    string
	RPDisplayName               string
	AttestationPreference       protocol.ConveyancePreference
	AuthenticatorSelection      protocol.AuthenticatorSelection
	MediationModalTimeout       int
	MediationConditionalTimeout int
}
