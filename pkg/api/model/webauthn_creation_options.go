package model

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
)

type WebAuthnCreationOptions struct {
	PublicKey PublicKeyCredentialCreationOptions      `json:"publicKey"`
	Mediation protocol.CredentialMediationRequirement `json:"mediation,omitempty"`
}

type PublicKeyCredentialCreationOptions struct {
	Challenge                     protocol.URLEncodedBase64           `json:"challenge"`
	RelyingParty                  PublicKeyCredentialRpEntity         `json:"rp"`
	User                          PublicKeyCredentialUserEntity       `json:"user"`
	PublicKeyCredentialParameters []PublicKeyCredentialParameter      `json:"pubKeyCredParams,omitempty"`
	Timeout                       int                                 `json:"timeout"`
	ExcludeCredentials            []PublicKeyCredentialDescriptor     `json:"excludeCredentials,omitempty"`
	AuthenticatorSelection        protocol.AuthenticatorSelection     `json:"authenticatorSelection"`
	Hints                         []protocol.PublicKeyCredentialHints `json:"hints,omitempty"`
	Attestation                   protocol.ConveyancePreference       `json:"attestation"`
	AttestationFormats            []protocol.AttestationFormat        `json:"attestationFormats,omitempty"`
	Extensions                    map[string]interface{}              `json:"extensions,omitempty"`
}

type PublicKeyCredentialRpEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublicKeyCredentialUserEntity struct {
	ID          protocol.URLEncodedBase64 `json:"id"`
	Name        string                    `json:"name"`
	DisplayName string                    `json:"displayName"`
}

type PublicKeyCredentialParameter struct {
	Type      protocol.CredentialType              `json:"type"`
	Algorithm webauthncose.COSEAlgorithmIdentifier `json:"alg"`
}

type PublicKeyCredentialDescriptor struct {
	Type       protocol.CredentialType   `json:"type"`
	ID         protocol.URLEncodedBase64 `json:"id"`
	Transports []string                  `json:"transports,omitempty"`
}
