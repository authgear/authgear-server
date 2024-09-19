package accountmanagement

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type ChangePasswordInput struct {
	OAuthSessionID string
	RedirectURI    string
	OldPassword    string
	NewPassword    string
}

type ChangePasswordOutput struct {
	RedirectURI string
}

// If have OAuthSessionID, it means the user is changing password after login with SDK.
// Then do special handling such as authenticationInfo
func (s *Service) ChangePassword(resolvedSession session.ResolvedSession, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	redirectURI := input.RedirectURI

	err := s.Database.WithTx(func() error {
		ais, err := s.Authenticators.List(
			userID,
			authenticator.KeepType(model.AuthenticatorTypePassword),
			authenticator.KeepKind(authenticator.KindPrimary),
		)
		if err != nil {
			return err
		}
		if len(ais) == 0 {
			return api.ErrNoPassword
		}
		oldInfo := ais[0]
		_, err = s.Authenticators.VerifyWithSpec(oldInfo, &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: input.OldPassword,
			},
		}, nil)
		if err != nil {
			err = api.ErrInvalidCredentials
			return err
		}
		changed, newInfo, err := s.Authenticators.UpdatePassword(oldInfo, &authenticatorservice.UpdatePasswordOptions{
			SetPassword:    true,
			PlainPassword:  input.NewPassword,
			SetExpireAfter: true,
		})
		if err != nil {
			return err
		}
		if changed {
			err = s.Authenticators.Update(newInfo)
			if err != nil {
				return err
			}
		}
		// If is changing password with SDK.
		if input.OAuthSessionID != "" {
			authInfo := resolvedSession.GetAuthenticationInfo()
			authenticationInfoEntry := authenticationinfo.NewEntry(authInfo, input.OAuthSessionID, "")

			err = s.AuthenticationInfoService.Save(authenticationInfoEntry)
			if err != nil {
				return err
			}
			redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(input.RedirectURI, authenticationInfoEntry)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ChangePasswordOutput{RedirectURI: redirectURI}, nil

}

type CreateAdditionalPasswordInput struct {
	PlainPassword string
}

func (s *Service) CreateAdditionalPassword(resolvedSession session.ResolvedSession, input CreateAdditionalPasswordInput) error {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: false,
		Kind:      model.AuthenticatorKindSecondary,
		Type:      model.AuthenticatorTypePassword,
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.PlainPassword,
		},
	}
	info, err := s.Authenticators.NewWithAuthenticatorID(uuid.New(), spec)
	if err != nil {
		return err
	}
	return s.createAuthenticator(info)
}

func (s *Service) createAuthenticator(authenticatorInfo *authenticator.Info) error {
	err := s.Database.WithTx(func() error {
		err := s.Authenticators.Create(authenticatorInfo, false)
		if err != nil {
			return err
		}
		if authenticatorInfo.Kind == authenticator.KindSecondary {
			err = s.Users.UpdateMFAEnrollment(authenticatorInfo.UserID, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
