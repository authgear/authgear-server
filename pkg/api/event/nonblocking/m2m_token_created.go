package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	M2MTokenCreated event.Type = "m2m.token.created" // #nosec G101
)

type M2MTokenCreatedEventPayload struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (e *M2MTokenCreatedEventPayload) UserID() string {
	return ""
}

func (e *M2MTokenCreatedEventPayload) NonBlockingEventType() event.Type {
	return M2MTokenCreated
}

func (e *M2MTokenCreatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *M2MTokenCreatedEventPayload) FillContext(ctx *event.Context) {
	ctx.ClientID = e.ClientID
}

func (e *M2MTokenCreatedEventPayload) ForHook() bool {
	return false
}

func (e *M2MTokenCreatedEventPayload) ForAudit() bool {
	return true
}

func (e *M2MTokenCreatedEventPayload) RequireReindexUserIDs() []string {
	return []string{}
}

func (e *M2MTokenCreatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &M2MTokenCreatedEventPayload{}
