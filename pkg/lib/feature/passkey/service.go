package passkey

import (
	"bytes"
	"encoding/base64"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Service struct {
	Store         *Store
	ConfigService *ConfigService
}

func (s *Service) PeekAttestationResponse(attestationResponse []byte) (creationOptions *model.WebAuthnCreationOptions, credentialID string, signCount int64, err error) {
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestationResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	session, err := s.Store.PeekSession(challenge)
	if err != nil {
		return
	}

	creationOptions = session.CreationOptions
	credentialID = parsed.ID
	signCount = int64(parsed.Response.AttestationObject.AuthData.Counter)
	return
}

func (s *Service) ConsumeAttestationResponse(attestationResponse []byte) (err error) {
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestationResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.ConsumeSession(challenge)
	if err != nil {
		return
	}

	return
}

func (s *Service) GetCredentialIDFromAssertionResponse(assertionResponse []byte) (credentialID string, err error) {
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(assertionResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.PeekSession(challenge)
	if err != nil {
		return
	}

	credentialID = parsed.ID
	return
}

func (s *Service) PeekAssertionResponse(assertionResponse []byte, attestationResponse []byte) (signCount int64, err error) {
	config, err := s.ConfigService.MakeConfig()
	if err != nil {
		return
	}

	parsedAssertion, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(assertionResponse))
	if err != nil {
		return
	}

	challengeString := parsedAssertion.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.PeekSession(challenge)
	if err != nil {
		return
	}

	parsedAttestation, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestationResponse))
	if err != nil {
		return
	}

	credential, err := webauthn.MakeNewCredential(parsedAttestation)
	if err != nil {
		return
	}

	err = parsedAssertion.Verify(
		challengeString,
		config.RPID,
		config.RPOrigin,
		"",    // We do not support FIDO AppID extension
		false, // user verification is preferred so we do not require user verification here.
		credential.PublicKey,
	)
	if err != nil {
		return
	}

	signCount = int64(parsedAssertion.Response.AuthenticatorData.Counter)
	return
}

func (s *Service) ConsumeAssertionResponse(assertionResponse []byte) (err error) {
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(assertionResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.ConsumeSession(challenge)
	if err != nil {
		return
	}

	return
}
