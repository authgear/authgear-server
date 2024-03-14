package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type ChangePasswordGetter interface {
	GetChangeReason() *interaction.AuthenticatorUpdateReason
}

type ChangePasswordViewModel struct {
	Force  bool
	Reason string
}

type ChangePasswordViewModeler struct {
	Authentication *config.AuthenticationConfig
	LoginID        *config.LoginIDConfig
}

func (m *ChangePasswordViewModeler) NewWithAuthflow(reason *declarative.PasswordChangeReason) ChangePasswordViewModel {
	var reasonStr string
	if reason != nil {
		reasonStr = string(*reason)
	}

	return ChangePasswordViewModel{
		Force:  true,
		Reason: reasonStr,
	}
}
