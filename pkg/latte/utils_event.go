package latte

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func DispatchAuthenticationFailedEvent(events workflow.EventService, info *authenticator.Info) error {
	return events.DispatchErrorEvent(&nonblocking.AuthenticationFailedEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: info.UserID,
			},
		},
		AuthenticationStage: string(info.Kind),
		AuthenticationType:  string(info.Type),
	})
}
