package passkey

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/go-webauthn/webauthn/protocol"

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

	// We used to call webauthn.MakeNewCredential(parsedAttestation) to obtain credentialBytes.
	// If you inspect the source code of that function, it is just doing some field mapping.
	// Therefore, we get rid of github.com/go-webauthn/webauthn/webauthn entirely, and pick the field from parsedAttestation directly.
	credentialBytes := parsedAttestation.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	// User verification is preferred so we do not require user verification here.
	verifyUser := false
	// webauthn prior to v0.13 always require the UP flag to be set.
	// Since v0.13 it allows UP flag to be unset.
	// See https://github.com/go-webauthn/webauthn/compare/v0.12.3...v0.13.4#diff-112f283f0b2011a522df0c3b95deea2464179fdebde208e15e99dba543f1fd19L399
	// To restore the previous behavior we have been using, we require the UP flag to be set.
	verifyUserPresence := true

	// Webauthn level 3 introduces topOrigin.
	// go-webauthn supports it since v0.11
	// Since we never expect Authgear to be iframe-ed, the list of expected top origins is the same as the list of expected origins.
	// Since we explicitly specify the list of expected top origins, we use protocol.TopOriginExplicitVerificationMode.
	// See https://www.w3.org/TR/webauthn-3/#dom-collectedclientdata-toporigin
	// See https://github.com/go-webauthn/webauthn/compare/v0.10.2...v0.11.2
	rpOrigins := []string{config.RPOrigin}
	rpTopOrigins := rpOrigins
	rpTopOriginVerify := protocol.TopOriginExplicitVerificationMode

	err = parsedAssertion.Verify(
		challengeString,
		config.RPID,
		rpOrigins,
		rpTopOrigins,
		rpTopOriginVerify,
		"", // We do not support FIDO AppID extension
		verifyUser,
		verifyUserPresence,
		credentialBytes,
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
