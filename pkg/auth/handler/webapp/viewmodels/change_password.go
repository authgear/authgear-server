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

func (m *ChangePasswordViewModeler) NewWithGraph(graph *interaction.Graph) ChangePasswordViewModel {
	var node ChangePasswordGetter
	if !graph.FindLastNode(&node) {
		panic("webapp: no node with password change reason found")
	}

	var reason string
	if node.GetChangeReason() != nil {
		reason = string(*node.GetChangeReason())
	}

	return ChangePasswordViewModel{
		Force:  true,
		Reason: reason,
	}
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
