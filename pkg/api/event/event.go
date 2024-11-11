package event

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type Type string

type StandardAttributesServiceNoEvent interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesServiceNoEvent interface {
	UpdateAllCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, reprForm map[string]interface{}) error
}

type RolesAndGroupsServiceNoEvent interface {
	ResetUserRole(ctx context.Context, options *rolesgroups.ResetUserRoleOptions) error
	ResetUserGroup(ctx context.Context, options *rolesgroups.ResetUserGroupOptions) error
}

type MutationsEffectContext struct {
	StandardAttributes StandardAttributesServiceNoEvent
	CustomAttributes   CustomAttributesServiceNoEvent
	RolesAndGroups     RolesAndGroupsServiceNoEvent
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
	ApplyMutations(ctx context.Context, mutations Mutations) bool
	// PerformEffects performs the side effects of the mutations.
	PerformEffects(ctx context.Context, effectCtx MutationsEffectContext) error
}

type NonBlockingPayload interface {
	Payload
	NonBlockingEventType() Type
	ForHook() bool
	ForAudit() bool
	RequireReindexUserIDs() []string
	DeletedUserIDs() []string
}

type Event struct {
	ID            string  `json:"id"`
	Seq           int64   `json:"seq"`
	Type          Type    `json:"type"`
	Payload       Payload `json:"payload"`
	Context       Context `json:"context"`
	IsNonBlocking bool    `json:"-"`
}

func (e *Event) ApplyMutations(ctx context.Context, mutations Mutations) bool {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		applied := blockingPayload.ApplyMutations(ctx, mutations)
		if applied {
			e.Payload = blockingPayload
			return true
		}
	}

	return false
}

func (e *Event) PerformEffects(ctx context.Context, effectCtx MutationsEffectContext) error {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		err := blockingPayload.PerformEffects(ctx, effectCtx)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
