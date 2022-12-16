package event

import (
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type Type string

type StandardAttributesServiceNoEvent interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesServiceNoEvent interface {
	UpdateAllCustomAttributes(role accesscontrol.Role, userID string, reprForm map[string]interface{}) error
}

type MutationsEffectContext struct {
	StandardAttributes StandardAttributesServiceNoEvent
	CustomAttributes   CustomAttributesServiceNoEvent
}

type Payload interface {
	UserID() string
	GetTriggeredBy() TriggeredByType
	FillContext(ctx *Context)
}

type BlockingPayload interface {
	Payload
	BlockingEventType() Type
	// ApplyMutations applies mutations to itself.
	ApplyMutations(mutations Mutations) bool
	// PerformEffects performs the side effects of the mutations.
	PerformEffects(ctx MutationsEffectContext) error
}

type NonBlockingPayload interface {
	Payload
	NonBlockingEventType() Type
	ForHook() bool
	ForAudit() bool
	ReindexUserNeeded() bool
	IsUserDeleted() bool
}

type Event struct {
	ID            string  `json:"id"`
	Seq           int64   `json:"seq"`
	Type          Type    `json:"type"`
	Payload       Payload `json:"payload"`
	Context       Context `json:"context"`
	IsNonBlocking bool    `json:"-"`
}

func (e *Event) ApplyMutations(mutations Mutations) bool {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		applied := blockingPayload.ApplyMutations(mutations)
		if applied {
			e.Payload = blockingPayload
			return true
		}
	}

	return false
}

func (e *Event) PerformEffects(ctx MutationsEffectContext) error {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		err := blockingPayload.PerformEffects(ctx)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
