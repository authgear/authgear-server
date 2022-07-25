package webauthn

import (
	"github.com/duo-labs/webauthn/protocol"
)

type RequestOptions struct {
	PublicKey PublicKeyCredentialRequestOptions `json:"publicKey"`
	Mediation string                            `json:"mediation,omitempty"`
}

type PublicKeyCredentialRequestOptions struct {
	Challenge        protocol.URLEncodedBase64            `json:"challenge"`
	Timeout          int                                  `json:"timeout"`
	RPID             string                               `json:"rpId"`
	UserVerification protocol.UserVerificationRequirement `json:"userVerification"`
	// This is a pointer to slice so that omitempty will omit the key if it is nil,
	// and it is an array if the value is non-nil.
	AllowCredentials *[]PublicKeyCredentialDescriptor `json:"allowCredentials,omitempty"`
	Extensions       map[string]interface{}           `json:"extensions,omitempty"`
}
