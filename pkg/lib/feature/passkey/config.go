package passkey

import (
	"github.com/go-webauthn/webauthn/protocol"
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
