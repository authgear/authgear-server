package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

// dispatchUserProfileUpdateIfFormFilled dispatches user.profile.pre_update and user.profile.updated
// once if any form was submitted in the flow.
//
// It searches flows.Nearest, not flows.Root: account linking nests an IntentLoginFlow in a signup flow,
// and that nested login must not count the forms of the outer signup, which were skipped.
func dispatchUserProfileUpdateIfFormFilled(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, userID string) error {
	filled := false
	for _, m := range authflow.FindAllMilestones[MilestoneDidFillForm](flows.Nearest) {
		if !m.MilestoneDidFillForm() {
			filled = true
			break
		}
	}
	if !filled {
		return nil
	}

	userRef := model.UserRef{Meta: model.Meta{ID: userID}}
	payloads := []event.Payload{
		&blocking.UserProfilePreUpdateBlockingEventPayload{UserRef: userRef},
		&nonblocking.UserProfileUpdatedEventPayload{UserRef: userRef},
	}
	for _, payload := range payloads {
		err := deps.Events.DispatchEventOnCommit(ctx, payload)
		if err != nil {
			return err
		}
	}
	return nil
}
