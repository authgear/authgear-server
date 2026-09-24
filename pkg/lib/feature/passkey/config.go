package passkey

import (
	"github.com/go-webauthn/webauthn/protocol"
)

type Config struct {
	RPID string
	// RPOrigins is every origin a ceremony for this project may report.
	// Verification compares the reported origin against these exactly, so an
	// origin that is missing here is rejected even when the browser considered
	// the ceremony legitimate.
	RPOrigins                   []string
	RPDisplayName               string
	AttestationPreference       protocol.ConveyancePreference
	AuthenticatorSelection      protocol.AuthenticatorSelection
	MediationModalTimeout       int
	MediationConditionalTimeout int
}
