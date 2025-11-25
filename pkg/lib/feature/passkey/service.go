package passkey

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Service struct {
	Store         *Store
	ConfigService *ConfigService
}

func (s *Service) PeekAttestationResponse(ctx context.Context, attestationResponse []byte) (creationOptions *model.WebAuthnCreationOptions, credentialID string, signCount int64, err error) {
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestationResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	session, err := s.Store.PeekSession(ctx, challenge)
	if err != nil {
		return
	}

	creationOptions = session.CreationOptions
	credentialID = parsed.ID
	signCount = int64(parsed.Response.AttestationObject.AuthData.Counter)
	return
}

func (s *Service) ConsumeAttestationResponse(ctx context.Context, attestationResponse []byte) (err error) {
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestationResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.ConsumeSession(ctx, challenge)
	if err != nil {
		return
	}

	return
}

func (s *Service) GetCredentialIDFromAssertionResponse(ctx context.Context, assertionResponse []byte) (credentialID string, err error) {
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(assertionResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.PeekSession(ctx, challenge)
	if err != nil {
		return
	}

	credentialID = parsed.ID
	return
}

func (s *Service) PeekAssertionResponse(ctx context.Context, assertionResponse []byte, attestationResponse []byte) (signCount int64, err error) {
	config, err := s.ConfigService.MakeConfig(ctx)
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

	_, err = s.Store.PeekSession(ctx, challenge)
	if err != nil {
		return
	}

	parsedAttestation, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestationResponse))
	if err != nil {
		return
	}

	// Compute the SHA-256 hash of the clientDataJSON for NewCredential
	clientDataJSON := parsedAttestation.Raw.AttestationResponse.ClientDataJSON
	h := sha256.Sum256(clientDataJSON)
	clientDataHash := h[:]

	credential, err := webauthn.NewCredential(clientDataHash, parsedAttestation)
	if err != nil {
		return
	}

	err = parsedAssertion.Verify(
		challengeString,
		config.RPID,
		[]string{config.RPOrigin},
		nil,                                      // rpTopOrigins - not using top-level origin verification
		protocol.TopOriginIgnoreVerificationMode, // Don't verify top-level origins
		"",                                       // We do not support FIDO AppID extension
		false,                                    // user verification is preferred so we do not require user verification here.
		true,                                     // require user presence
		credential.PublicKey,
	)
	if err != nil {
		return
	}

	signCount = int64(parsedAssertion.Response.AuthenticatorData.Counter)
	return
}

func (s *Service) ConsumeAssertionResponse(ctx context.Context, assertionResponse []byte) (err error) {
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(assertionResponse))
	if err != nil {
		return
	}

	challengeString := parsed.Response.CollectedClientData.Challenge
	challenge, err := base64.RawURLEncoding.DecodeString(challengeString)
	if err != nil {
		return
	}

	_, err = s.Store.ConsumeSession(ctx, challenge)
	if err != nil {
		return
	}

	return
}
