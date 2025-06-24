package event

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
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

type RateLimiter interface {
	AdjustWeight(ctx context.Context, r *ratelimit.Reservation, weight float64) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
}

type MutationsEffectContext struct {
	StandardAttributes StandardAttributesServiceNoEvent
	CustomAttributes   CustomAttributesServiceNoEvent
	RolesAndGroups     RolesAndGroupsServiceNoEvent
}

type RateLimitContext struct {
	RateLimiter RateLimiter
}

type Payload interface {
	UserID() string
	GetTriggeredBy() TriggeredByType
	FillContext(ctx *Context)
}

type ApplyHookResponseResult struct {
	MutationsEverApplied bool
	RateLimitEverApplied bool
}

type BlockingPayload interface {
	Payload
	BlockingEventType() Type
	// ApplyHookResponse applies hook response to itself.
	ApplyHookResponse(ctx context.Context, response HookResponse) ApplyHookResponseResult
	// PerformMutationEffects performs the side effects of the mutations.
	PerformMutationEffects(ctx context.Context, eventCtx Context, effectCtx MutationsEffectContext) error
	// PerformRateLimitEffects performs the side effects of rate_limit.
	PerformRateLimitEffects(ctx context.Context, eventCtx Context, effectCtx RateLimitContext) error
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

func (e *Event) ApplyHookResponse(ctx context.Context, response HookResponse) ApplyHookResponseResult {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		result := blockingPayload.ApplyHookResponse(ctx, response)
		if result.MutationsEverApplied {
			e.Payload = blockingPayload
		}
		return result
	}

	return ApplyHookResponseResult{}
}

func (e *Event) PerformMutationsEffects(ctx context.Context, effectCtx MutationsEffectContext) error {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		err := blockingPayload.PerformMutationEffects(ctx, e.Context, effectCtx)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (e *Event) PerformRateLimitEffects(ctx context.Context, effectCtx RateLimitContext) error {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		err := blockingPayload.PerformRateLimitEffects(ctx, e.Context, effectCtx)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
