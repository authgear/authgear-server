package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SettingsViewModel struct {
	Authenticators                  []*authenticator.Info
	SecondaryAuthenticationDisabled bool
	SecondaryTOTPAllowed            bool
	SecondaryOOBOTPEmailAllowed     bool
	SecondaryOOBOTPSMSAllowed       bool
	SecondaryPasswordAllowed        bool
	HasDeviceTokens                 bool
	ListRecoveryCodesAllowed        bool
	ShowBiometric                   bool
}

type SettingsIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type SettingsAuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsMFAService interface {
	HasDeviceTokens(userID string) (bool, error)
}

type SettingsViewModeler struct {
	Authenticators SettingsAuthenticatorService
	MFA            SettingsMFAService
	Authentication *config.AuthenticationConfig
	Biometric      *config.BiometricConfig
}

func (m *SettingsViewModeler) ViewModel(userID string) (*SettingsViewModel, error) {
	authenticators, err := m.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}

	somePrimaryAuthenticatorCanHaveMFA := len(authenticator.ApplyFilters(
		authenticators,
		authenticator.KeepPrimaryAuthenticatorCanHaveMFA,
	)) > 0

	hasDeviceTokens, err := m.MFA.HasDeviceTokens(userID)
	if err != nil {
		return nil, err
	}

	totp := false
	oobotpemail := false
	oobotpsms := false
	password := false

	if somePrimaryAuthenticatorCanHaveMFA {
		for _, typ := range *m.Authentication.SecondaryAuthenticators {
			switch typ {
			case model.AuthenticatorTypePassword:
				password = true
			case model.AuthenticatorTypeTOTP:
				totp = true
			case model.AuthenticatorTypeOOBEmail:
				oobotpemail = true
			case model.AuthenticatorTypeOOBSMS:
				oobotpsms = true
			}
		}
	}

	showBiometric := false
	for _, typ := range m.Authentication.Identities {
		if typ == model.IdentityTypeBiometric && *m.Biometric.ListEnabled {
			showBiometric = true
		}
	}

	viewModel := &SettingsViewModel{
		Authenticators:                  authenticators,
		SecondaryAuthenticationDisabled: m.Authentication.SecondaryAuthenticationMode.IsDisabled(),
		SecondaryTOTPAllowed:            totp,
		SecondaryOOBOTPEmailAllowed:     oobotpemail,
		SecondaryOOBOTPSMSAllowed:       oobotpsms,
		SecondaryPasswordAllowed:        password,
		HasDeviceTokens:                 hasDeviceTokens,
		ListRecoveryCodesAllowed:        !*m.Authentication.RecoveryCode.Disabled && m.Authentication.RecoveryCode.ListEnabled,
		ShowBiometric:                   showBiometric,
	}
	return viewModel, nil
}
