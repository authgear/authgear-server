package passkey

import (
	"encoding/base64"

	"github.com/duo-labs/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type RequestOptionsService struct {
	ConfigService   *ConfigService
	IdentityService IdentityService
	Store           *Store
}

func (s *RequestOptionsService) MakeConditionalRequestOptions() (*model.WebAuthnRequestOptions, error) {
	challenge, err := protocol.CreateChallenge()
	if err != nil {
		return nil, err
	}

	config, err := s.ConfigService.MakeConfig()
	if err != nil {
		return nil, err
	}

	options := &model.WebAuthnRequestOptions{
		Mediation: "conditional",
		PublicKey: model.PublicKeyCredentialRequestOptions{
			Challenge:        challenge,
			Timeout:          config.MediationConditionalTimeout,
			RPID:             config.RPID,
			UserVerification: config.AuthenticatorSelection.UserVerification,
			// Conditional request is for AutoFill.
			// So any credential that exists on the platform is allowed.
			AllowCredentials: nil,
			Extensions: map[string]interface{}{
				// We want to know user verification method (uvm).
				// https://www.w3.org/TR/webauthn-2/#sctn-uvm-extension
				"uvm": true,
			},
		},
	}

	session := &Session{
		Challenge:      challenge,
		RequestOptions: options,
	}
	err = s.Store.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return options, nil
}

func (s *RequestOptionsService) MakeModalRequestOptions(userID string) (*model.WebAuthnRequestOptions, error) {
	challenge, err := protocol.CreateChallenge()
	if err != nil {
		return nil, err
	}

	config, err := s.ConfigService.MakeConfig()
	if err != nil {
		return nil, err
	}

	options := &model.WebAuthnRequestOptions{
		PublicKey: model.PublicKeyCredentialRequestOptions{
			Challenge:        challenge,
			Timeout:          config.MediationModalTimeout,
			RPID:             config.RPID,
			UserVerification: config.AuthenticatorSelection.UserVerification,
			Extensions: map[string]interface{}{
				// We want to know user verification method (uvm).
				// https://www.w3.org/TR/webauthn-2/#sctn-uvm-extension
				"uvm": true,
			},
		},
	}

	// Populate AllowCredentials
	// Make it an array so that if the user has no passkey,
	// allowCredentials is an empty array.
	// Thus the platform will disallow the user from selecting anything.
	allow := []model.PublicKeyCredentialDescriptor{}
	identities, err := s.IdentityService.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range identities {
		if i.Type == model.IdentityTypePasskey {
			credentialID := i.Passkey.CredentialID
			credentialIDBytes, err := base64.RawStdEncoding.DecodeString(credentialID)
			if err != nil {
				return nil, err
			}
			allow = append(allow, model.PublicKeyCredentialDescriptor{
				Type: protocol.PublicKeyCredentialType,
				ID:   protocol.URLEncodedBase64(credentialIDBytes),
			})
		}
	}
	options.PublicKey.AllowCredentials = &allow

	session := &Session{
		Challenge:      challenge,
		RequestOptions: options,
	}
	err = s.Store.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return options, nil
}
