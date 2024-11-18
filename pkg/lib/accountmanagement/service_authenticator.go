package accountmanagement

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type ChangePrimaryPasswordInput struct {
	OAuthSessionID string
	RedirectURI    string
	OldPassword    string
	NewPassword    string
}

type ChangePrimaryPasswordOutput struct {
	RedirectURI string
}

// If have OAuthSessionID, it means the user is changing password after login with SDK.
// Then do special handling such as authenticationInfo
func (s *Service) ChangePrimaryPassword(resolvedSession session.ResolvedSession, input *ChangePrimaryPasswordInput) (*ChangePrimaryPasswordOutput, error) {
	redirectURI := input.RedirectURI

	var err error
	err = s.Database.WithTx(func() error {
		_, err = s.changePassword(resolvedSession, &changePasswordInput{
			Kind:        authenticator.KindPrimary,
			OldPassword: input.OldPassword,
			NewPassword: input.NewPassword,
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// If is changing password with SDK.
	if input.OAuthSessionID != "" {
		authInfo := resolvedSession.GetAuthenticationInfo()
		authenticationInfoEntry := authenticationinfo.NewEntry(authInfo, input.OAuthSessionID, "")

		err = s.AuthenticationInfoService.Save(authenticationInfoEntry)
		if err != nil {
			return nil, err
		}
		redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(input.RedirectURI, authenticationInfoEntry)
	}

	return &ChangePrimaryPasswordOutput{
		RedirectURI: redirectURI,
	}, nil
}

type CreateSecondaryPasswordInput struct {
	PlainPassword string
}

type CreateSecondaryPasswordOutput struct {
}

func (s *Service) CreateSecondaryPassword(resolvedSession session.ResolvedSession, input CreateSecondaryPasswordInput) (*CreateSecondaryPasswordOutput, error) {
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
	info, err := s.Authenticators.New(spec)
	if err != nil {
		return nil, err
	}
	err = s.Database.WithTx(func() error {
		return s.createAuthenticator(info)
	})
	if err != nil {
		return nil, err
	}
	return &CreateSecondaryPasswordOutput{}, nil
}

type ChangeSecondaryPasswordInput struct {
	OldPassword string
	NewPassword string
}

type ChangeSecondaryPasswordOutput struct {
}

func (s *Service) ChangeSecondaryPassword(resolvedSession session.ResolvedSession, input *ChangeSecondaryPasswordInput) (*ChangeSecondaryPasswordOutput, error) {
	err := s.Database.WithTx(func() error {
		_, err := s.changePassword(resolvedSession, &changePasswordInput{
			Kind:        authenticator.KindSecondary,
			OldPassword: input.OldPassword,
			NewPassword: input.NewPassword,
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ChangeSecondaryPasswordOutput{}, nil
}

type DeleteSecondaryPasswordInput struct {
}

type DeleteSecondaryPasswordOutput struct {
}

func (s *Service) DeleteSecondaryPassword(resolvedSession session.ResolvedSession, input *DeleteSecondaryPasswordInput) (*DeleteSecondaryPasswordOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	err := s.Database.WithTx(func() error {
		info, err := s.prepareDeleteSecondaryPassword(userID)
		if err != nil {
			return err
		}
		info, err = s.prepareDeleteAuthenticator(userID, info.ID)
		if err != nil {
			return err
		}

		err = s.Authenticators.Delete(info)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DeleteSecondaryPasswordOutput{}, nil
}

func (s *Service) prepareDeleteSecondaryPassword(userID string) (*authenticator.Info, error) {
	ais, err := s.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(authenticator.KindSecondary),
	)
	if err != nil {
		return nil, err
	}
	if len(ais) == 0 {
		return nil, api.ErrNoPassword
	}
	return ais[0], nil
}

type StartAddTOTPAuthenticatorInput struct{}
type StartAddTOTPAuthenticatorOutput struct {
	Token                   string
	EndUserAccountID        string
	AuthenticatorTOTPIssuer string
	AuthenticatorTOTPSecret string
}

func (s *Service) StartAddTOTPAuthenticator(resolvedSession session.ResolvedSession, input *StartAddTOTPAuthenticatorInput) (*StartAddTOTPAuthenticatorOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	var endUserAccountID string
	err := s.Database.WithTx(func() error {
		user, err := s.Users.Get(userID, accesscontrol.RoleGreatest)
		if err != nil {
			return err
		}
		endUserAccountID = user.EndUserAccountID
		return nil
	})
	if err != nil {
		return nil, err
	}

	totp, err := secretcode.NewTOTPFromRNG()
	if err != nil {
		return nil, err
	}
	token, err := s.Store.GenerateToken(GenerateTokenOptions{
		UserID:                            userID,
		AuthenticatorType:                 model.AuthenticatorTypeTOTP,
		AuthenticatorTOTPIssuer:           string(s.HTTPOrigin),
		AuthenticatorTOTPEndUserAccountID: endUserAccountID,
		AuthenticatorTOTPSecret:           totp.Secret,
	})
	if err != nil {
		return nil, err
	}

	return &StartAddTOTPAuthenticatorOutput{
		Token:                   token,
		EndUserAccountID:        endUserAccountID,
		AuthenticatorTOTPIssuer: string(s.HTTPOrigin),
		AuthenticatorTOTPSecret: totp.Secret,
	}, nil
}

type ResumeAddTOTPAuthenticatorInput struct {
	DisplayName string
	Code        string
}
type ResumeAddTOTPAuthenticatorOutput struct {
	Token                string
	RecoveryCodesCreated bool
}

func (s *Service) ResumeAddTOTPAuthenticator(resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddTOTPAuthenticatorInput) (output *ResumeAddTOTPAuthenticatorOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(tokenString)
		}
	}()

	if err != nil {
		return
	}

	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	info, err := s.Authenticators.New(
		&authenticator.Spec{
			UserID:    userID,
			IsDefault: false,
			Kind:      model.AuthenticatorKindSecondary,
			Type:      model.AuthenticatorTypeTOTP,
			TOTP: &authenticator.TOTPSpec{
				DisplayName: input.DisplayName,
				Secret:      token.Authenticator.TOTPSecret,
			},
		},
	)
	if err != nil {
		return
	}
	_, err = s.Authenticators.VerifyWithSpec(
		info,
		&authenticator.Spec{
			UserID:    userID,
			IsDefault: false,
			Kind:      model.AuthenticatorKindSecondary,
			Type:      model.AuthenticatorTypeTOTP,
			TOTP: &authenticator.TOTPSpec{
				DisplayName: input.DisplayName,
				Code:        input.Code,
			},
		},
		nil,
	)

	if err != nil {
		return
	}

	recoveryCodes, recoveryCodesCreated, err := s.generateRecoveryCodes(userID)

	newToken, err := s.Store.GenerateToken(GenerateTokenOptions{
		UserID:                            userID,
		AuthenticatorRecoveryCodes:        recoveryCodes,
		AuthenticatorRecoveryCodesCreated: recoveryCodesCreated,
		AuthenticatorType:                 model.AuthenticatorType(token.Authenticator.AuthenticatorType),
		AuthenticatorTOTPDisplayName:      input.DisplayName,
		AuthenticatorTOTPSecret:           token.Authenticator.TOTPSecret,
		AuthenticatorTOTPVerified:         true,
	})
	if err != nil {
		return
	}

	output = &ResumeAddTOTPAuthenticatorOutput{
		Token:                newToken,
		RecoveryCodesCreated: recoveryCodesCreated,
	}
	return
}

type FinishAddTOTPAuthenticatorInput struct {
}

type FinishAddTOTPAuthenticatorOutput struct {
	Info *authenticator.Info
}

func (s *Service) FinishAddTOTPAuthenticator(resolvedSession session.ResolvedSession, tokenString string, input *FinishAddTOTPAuthenticatorInput) (output *FinishAddTOTPAuthenticatorOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(tokenString)
		}
	}()

	if err != nil {
		return
	}

	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	info, err := s.Authenticators.New(
		&authenticator.Spec{
			UserID:    userID,
			IsDefault: false,
			Kind:      model.AuthenticatorKindSecondary,
			Type:      model.AuthenticatorTypeTOTP,
			TOTP: &authenticator.TOTPSpec{
				DisplayName: token.Authenticator.TOTPDisplayName,
				Secret:      token.Authenticator.TOTPSecret,
			},
		},
	)
	if err != nil {
		return
	}
	err = s.Database.WithTx(func() error {
		err = s.createAuthenticator(info)
		if err != nil {
			return err
		}

		if token.Authenticator.RecoveryCodesCreated {
			_, err = s.MFA.ReplaceRecoveryCodes(userID, token.Authenticator.RecoveryCodes)
			if err != nil {
				return err
			}
		}

		return nil
	})

	output = &FinishAddTOTPAuthenticatorOutput{
		Info: info,
	}
	return
}

type DeleteTOTPAuthenticatorInput struct {
	AuthenticatorID string
}

type DeleteTOTPAuthenticatorOutput struct {
	Info *authenticator.Info
}

func (s *Service) DeleteTOTPAuthenticator(resolvedSession session.ResolvedSession, input *DeleteTOTPAuthenticatorInput) (output *DeleteTOTPAuthenticatorOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	authenticatorID := input.AuthenticatorID

	var info *authenticator.Info
	err = s.Database.WithTx(func() error {
		info, err = s.prepareDeleteAuthenticator(userID, authenticatorID)
		if err != nil {
			return err
		}

		err = s.Authenticators.Delete(info)
		if err != nil {
			return err
		}

		return nil
	})

	output = &DeleteTOTPAuthenticatorOutput{
		Info: info,
	}
	return
}

func (s *Service) prepareDeleteAuthenticator(userID string, authenticatorID string) (*authenticator.Info, error) {
	info, err := s.Authenticators.Get(authenticatorID)
	if err != nil {
		return nil, err
	}

	if info.UserID != userID {
		return nil, ErrAccountManagementAuthenticatorNotOwnedbyToUser
	}

	// Return error if secondary authentication is required,
	// and this is the only secondary authenticator the user has
	if s.Config.Authentication.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeRequired {
		otherSecondaryAuthns, err := s.Authenticators.List(
			userID,
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.FilterFunc(func(ai *authenticator.Info) bool {
				return ai.ID != info.ID
			}),
		)
		if err != nil {
			return nil, err
		}
		if len(otherSecondaryAuthns) == 0 {
			return nil, ErrAccountManagementSecondaryAuthenticatorIsRequired
		}
	}

	return info, nil
}

type changePasswordInput struct {
	Kind        authenticator.Kind
	OldPassword string
	NewPassword string
}

type changePasswordOutput struct {
}

func (s *Service) changePassword(resolvedSession session.ResolvedSession, input *changePasswordInput) (*changePasswordOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	ais, err := s.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(input.Kind),
	)
	if err != nil {
		return nil, err
	}
	if len(ais) == 0 {
		return nil, api.ErrNoPassword
	}
	oldInfo := ais[0]
	_, err = s.Authenticators.VerifyWithSpec(oldInfo, &authenticator.Spec{
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.OldPassword,
		},
	}, nil)
	if err != nil {
		err = api.ErrInvalidCredentials
		return nil, err
	}
	changed, newInfo, err := s.Authenticators.UpdatePassword(oldInfo, &authenticatorservice.UpdatePasswordOptions{
		SetPassword:    true,
		PlainPassword:  input.NewPassword,
		SetExpireAfter: true,
	})
	if err != nil {
		return nil, err
	}
	if changed {
		err = s.Authenticators.Update(newInfo)
		if err != nil {
			return nil, err
		}

		// switch input.Kind {
		// case authenticator.KindPrimary:
		// 	err = s.Events.DispatchEventOnCommit(ctx, &nonblocking.PasswordPrimaryChangedEventPayload{
		// 		UserRef: model.UserRef{
		// 			Meta: model.Meta{
		// 				ID: userID,
		// 			},
		// 		},
		// 	})
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// case authenticator.KindSecondary:
		// 	err = s.Events.DispatchEventOnCommit(ctx, &nonblocking.PasswordSecondaryChangedEventPayload{
		// 		UserRef: model.UserRef{
		// 			Meta: model.Meta{
		// 				ID: userID,
		// 			},
		// 		},
		// 	})
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// default:
		// 	panic(fmt.Errorf("unexpected authenticator kind: %v", input.Kind))
		// }
	}

	return &changePasswordOutput{}, nil
}

func (s *Service) createAuthenticator(authenticatorInfo *authenticator.Info) error {
	err := s.Authenticators.Create(authenticatorInfo, true)
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
}

type StartAddOOBOTPAuthenticatorInput struct {
	Channel model.AuthenticatorOOBChannel
	Target  string
}
type StartAddOOBOTPAuthenticatorOutput struct {
	Token string
}

func (s *Service) StartAddOOBOTPAuthenticator(resolvedSession session.ResolvedSession, input *StartAddOOBOTPAuthenticatorInput) (*StartAddOOBOTPAuthenticatorOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	_, err := s.prepareNewAuthenticator(userID, input.Channel, input.Target)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() error {
		err := s.sendOTPCode(userID, input.Channel, input.Target, false)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var authenticatorType model.AuthenticatorType
	switch input.Channel {
	case model.AuthenticatorOOBChannelWhatsapp:
		fallthrough
	case model.AuthenticatorOOBChannelSMS:
		authenticatorType = model.AuthenticatorTypeOOBSMS
	case model.AuthenticatorOOBChannelEmail:
		authenticatorType = model.AuthenticatorTypeOOBEmail
	default:
		panic("unexpected channel")
	}

	token, err := s.Store.GenerateToken(GenerateTokenOptions{
		UserID:                     userID,
		AuthenticatorType:          authenticatorType,
		AuthenticatorOOBOTPChannel: input.Channel,
		AuthenticatorOOBOTPTarget:  input.Target,
	})
	if err != nil {
		return nil, err
	}

	return &StartAddOOBOTPAuthenticatorOutput{
		Token: token,
	}, nil
}

type ResumeAddOOBOTPAuthenticatorInput struct {
	Code string
}
type ResumeAddOOBOTPAuthenticatorOutput struct {
	Token                string
	RecoveryCodesCreated bool
}

func (s *Service) ResumeAddOOBOTPAuthenticator(resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddOOBOTPAuthenticatorInput) (output *ResumeAddOOBOTPAuthenticatorOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(tokenString)
		}
	}()

	if err != nil {
		return
	}

	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.VerifyOTP(
		userID,
		token.Authenticator.OOBOTPChannel,
		token.Authenticator.OOBOTPTarget,
		input.Code,
		false,
	)
	if err != nil {
		return
	}

	recoveryCodes, recoveryCodesCreated, err := s.generateRecoveryCodes(userID)

	newToken, err := s.Store.GenerateToken(GenerateTokenOptions{
		UserID:                            userID,
		AuthenticatorRecoveryCodes:        recoveryCodes,
		AuthenticatorRecoveryCodesCreated: recoveryCodesCreated,
		AuthenticatorType:                 model.AuthenticatorType(token.Authenticator.AuthenticatorType),
		AuthenticatorOOBOTPChannel:        token.Authenticator.OOBOTPChannel,
		AuthenticatorOOBOTPTarget:         token.Authenticator.OOBOTPTarget,
		AuthenticatorOOBOTPVerified:       true,
	})
	if err != nil {
		return
	}

	output = &ResumeAddOOBOTPAuthenticatorOutput{
		Token:                newToken,
		RecoveryCodesCreated: recoveryCodesCreated,
	}
	return
}

type FinishAddOOBOTPAuthenticatorInput struct {
}

type FinishAddOOBOTPAuthenticatorOutput struct {
	Info *authenticator.Info
}

func (s *Service) FinishAddOOBOTPAuthenticator(resolvedSession session.ResolvedSession, tokenString string, input *FinishAddOOBOTPAuthenticatorInput) (output *FinishAddOOBOTPAuthenticatorOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(tokenString)
		}
	}()

	if err != nil {
		return
	}

	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	info, err := s.prepareNewAuthenticator(userID, token.Authenticator.OOBOTPChannel, token.Authenticator.OOBOTPTarget)
	if err != nil {
		return
	}

	err = s.Database.WithTx(func() error {
		err = s.createAuthenticator(info)
		if err != nil {
			return err
		}

		if token.Authenticator.RecoveryCodesCreated {
			_, err = s.MFA.ReplaceRecoveryCodes(userID, token.Authenticator.RecoveryCodes)
			if err != nil {
				return err
			}
		}

		return nil
	})

	output = &FinishAddOOBOTPAuthenticatorOutput{
		Info: info,
	}
	return
}

func (s *Service) prepareNewAuthenticator(userID string, channel model.AuthenticatorOOBChannel, target string) (*authenticator.Info, error) {
	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: false,
		Kind:      model.AuthenticatorKindSecondary,
		OOBOTP:    &authenticator.OOBOTPSpec{},
	}

	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		spec.Type = model.AuthenticatorTypeOOBEmail
		spec.OOBOTP.Email = target
	case model.AuthenticatorOOBChannelWhatsapp:
		fallthrough
	case model.AuthenticatorOOBChannelSMS:
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = target
	default:
		panic("unexpected channel")
	}

	info, err := s.Authenticators.New(spec)
	if err != nil {
		return nil, err
	}

	return info, nil
}

type DeleteOOBOTPAuthenticatorInput struct {
	AuthenticatorID string
}

type DeleteOOBOTPAuthenticatorOutput struct {
	Info *authenticator.Info
}

func (s *Service) DeleteOOBOTPAuthenticator(resolvedSession session.ResolvedSession, input *DeleteOOBOTPAuthenticatorInput) (output *DeleteOOBOTPAuthenticatorOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	authenticatorID := input.AuthenticatorID

	var info *authenticator.Info
	err = s.Database.WithTx(func() error {
		info, err = s.prepareDeleteAuthenticator(userID, authenticatorID)
		if err != nil {
			return err
		}

		err = s.Authenticators.Delete(info)
		if err != nil {
			return err
		}

		return nil
	})

	output = &DeleteOOBOTPAuthenticatorOutput{
		Info: info,
	}
	return
}

func (s *Service) generateRecoveryCodes(userID string) (recoveryCodes []string, isCreated bool, err error) {
	if *s.Config.Authentication.RecoveryCode.Disabled {
		return nil, false, nil
	}

	err = s.Database.WithTx(func() error {
		existing, err := s.MFA.ListRecoveryCodes(userID)
		if err != nil {
			return err
		}

		if len(existing) == 0 {
			isCreated = true
			recoveryCodes = s.MFA.GenerateRecoveryCodes()
			return nil
		}

		return nil
	})

	return recoveryCodes, isCreated, err
}

type GenerateRecoveryCodesInput struct {
}

type GenerateRecoveryCodesOutput struct {
	Info *authenticator.Info
}

func (s *Service) GenerateRecoveryCodes(resolvedSession session.ResolvedSession, input *GenerateRecoveryCodesInput) (output *GenerateRecoveryCodesOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	recoveryCodes := s.MFA.GenerateRecoveryCodes()
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() error {
		_, err = s.MFA.ReplaceRecoveryCodes(userID, recoveryCodes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	output = &GenerateRecoveryCodesOutput{}
	return output, nil
}
