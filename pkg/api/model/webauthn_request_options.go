package model

import (
	"github.com/go-webauthn/webauthn/protocol"
)

type WebAuthnRequestOptions struct {
	PublicKey PublicKeyCredentialRequestOptions       `json:"publicKey"`
	Mediation protocol.CredentialMediationRequirement `json:"mediation,omitempty"`
}

type PublicKeyCredentialRequestOptions struct {
	Challenge        protocol.URLEncodedBase64            `json:"challenge"`
	Timeout          int                                  `json:"timeout"`
	RPID             string                               `json:"rpId"`
	UserVerification protocol.UserVerificationRequirement `json:"userVerification"`
	// This is a pointer to slice so that omitempty will omit the key if it is nil,
	// and it is an array if the value is non-nil.
	AllowCredentials *[]PublicKeyCredentialDescriptor    `json:"allowCredentials,omitempty"`
	Hints            []protocol.PublicKeyCredentialHints `json:"hints,omitempty"`
	Extensions       map[string]interface{}              `json:"extensions,omitempty"`
}
