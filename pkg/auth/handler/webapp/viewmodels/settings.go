package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
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
	ShowBiometric               bool
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
	Identities     SettingsIdentityService
	MFA            SettingsMFAService
	Authentication *config.AuthenticationConfig
	Biometric      *config.BiometricConfig
}

func (m *SettingsViewModeler) ViewModel(userID string) (*SettingsViewModel, error) {
	authenticators, err := m.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}

	iis, err := m.Identities.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	someIdentityCanHaveMFA := false
	for _, ii := range iis {
		if ii.CanHaveMFA() {
			someIdentityCanHaveMFA = true
		}
	}

	hasDeviceTokens, err := m.MFA.HasDeviceTokens(userID)
	if err != nil {
		return nil, err
	}

	totp := false
	oobotpemail := false
	oobotpsms := false
	password := false

	if someIdentityCanHaveMFA {
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
	}

	showBiometric := false
	for _, typ := range m.Authentication.Identities {
		if typ == authn.IdentityTypeBiometric && *m.Biometric.ListEnabled {
			showBiometric = true
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
		ShowBiometric:               showBiometric,
	}
	return viewModel, nil
}
