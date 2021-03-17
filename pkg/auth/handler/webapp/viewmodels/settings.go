package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SettingsViewModel struct {
	Authenticators              []*authenticator.Info
	SecondaryTOTPAllowed        bool
	SecondaryOOBOTPEmailAllowed bool
	SecondaryOOBOTPSMSAllowed   bool
	SecondaryPasswordAllowed    bool
	HasDeviceTokens             bool
	ListRecoveryCodesAllowed    bool
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
}

func (m *SettingsViewModeler) ViewModel(userID string) (*SettingsViewModel, error) {
	authenticators, err := m.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}

	hasDeviceTokens, err := m.MFA.HasDeviceTokens(userID)
	if err != nil {
		return nil, err
	}

	totp := false
	oobotpemail := false
	oobotpsms := false
	password := false
	for _, typ := range m.Authentication.SecondaryAuthenticators {
		switch typ {
		case authn.AuthenticatorTypePassword:
			password = true
		case authn.AuthenticatorTypeTOTP:
			totp = true
		case authn.AuthenticatorTypeOOBEmail:
			oobotpemail = true
		case authn.AuthenticatorTypeOOBSMS:
			oobotpsms = true
		}
	}
	viewModel := &SettingsViewModel{
		Authenticators:              authenticators,
		SecondaryTOTPAllowed:        totp,
		SecondaryOOBOTPEmailAllowed: oobotpemail,
		SecondaryOOBOTPSMSAllowed:   oobotpsms,
		SecondaryPasswordAllowed:    password,
		HasDeviceTokens:             hasDeviceTokens,
		ListRecoveryCodesAllowed:    m.Authentication.RecoveryCode.ListEnabled,
	}
	return viewModel, nil
}
