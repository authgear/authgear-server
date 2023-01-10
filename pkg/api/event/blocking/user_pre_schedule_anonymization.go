package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreScheduleAnonymization event.Type = "user.pre_schedule_anonymization"
)

type UserPreScheduleAnonymizationBlockingEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) BlockingEventType() event.Type {
	return UserPreScheduleAnonymization
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) ApplyMutations(mutations event.Mutations) bool {
	user, mutated := ApplyUserMutations(e.UserModel, mutations.User)
	if mutated {
		e.UserModel = user
		return true
	}

	return false
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) PerformEffects(ctx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, userID, userMutations)
}

var _ event.BlockingPayload = &UserPreScheduleAnonymizationBlockingEventPayload{}
