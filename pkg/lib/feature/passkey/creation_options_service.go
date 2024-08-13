package passkey

import (
	"encoding/base64"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type UserService interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
}

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type CreationOptionsService struct {
	ConfigService   *ConfigService
	UserService     UserService
	IdentityService IdentityService
	Store           *Store
}

// MakeCreationOptions makes creation options which is ready for use.
func (s *CreationOptionsService) MakeCreationOptions(userID string) (*model.WebAuthnCreationOptions, error) {
	challenge, err := protocol.CreateChallenge()
	if err != nil {
		return nil, err
	}

	config, err := s.ConfigService.MakeConfig()
	if err != nil {
		return nil, err
	}

	user, err := s.UserService.Get(userID, accesscontrol.RoleGreatest)
	if err != nil {
		return nil, err
	}

	endUserAccountID := user.EndUserAccountID

	var exclude []model.PublicKeyCredentialDescriptor
	identities, err := s.IdentityService.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range identities {
		if i.Type == model.IdentityTypePasskey {
			credentialID := i.Passkey.CredentialID
			credentialIDBytes, err := base64.RawURLEncoding.DecodeString(credentialID)
			if err != nil {
				return nil, err
			}
			exclude = append(exclude, model.PublicKeyCredentialDescriptor{
				Type: protocol.PublicKeyCredentialType,
				ID:   protocol.URLEncodedBase64(credentialIDBytes),
			})
		}
	}

	options := &model.WebAuthnCreationOptions{
		PublicKey: model.PublicKeyCredentialCreationOptions{
			Challenge: challenge,
			RelyingParty: model.PublicKeyCredentialRpEntity{
				ID:   config.RPID,
				Name: config.RPDisplayName,
			},
			User: model.PublicKeyCredentialUserEntity{
				ID:          []byte(user.ID),
				Name:        endUserAccountID,
				DisplayName: endUserAccountID,
			},
			// https://www.w3.org/TR/webauthn-2/#CreateCred-DetermineRpId
			// The default in the spec is ES256 and RS256.
			PublicKeyCredentialParameters: []model.PublicKeyCredentialParameter{
				{
					Type:      protocol.PublicKeyCredentialType,
					Algorithm: webauthncose.AlgES256,
				},
				{
					Type:      protocol.PublicKeyCredentialType,
					Algorithm: webauthncose.AlgRS256,
				},
			},
			Extensions: map[string]interface{}{
				// We want to know user verification method (uvm).
				// https://www.w3.org/TR/webauthn-2/#sctn-uvm-extension
				"uvm": true,
				// We want to know the credentials is client-side discoverable or not.
				// https://www.w3.org/TR/webauthn-2/#sctn-authenticator-credential-properties-extension
				"credProps": true,
			},
			AuthenticatorSelection: config.AuthenticatorSelection,
			Timeout:                config.MediationModalTimeout,
			Attestation:            config.AttestationPreference,
			ExcludeCredentials:     exclude,
		},
	}

	session := &Session{
		Challenge:       challenge,
		CreationOptions: options,
	}

	err = s.Store.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return options, nil
}
