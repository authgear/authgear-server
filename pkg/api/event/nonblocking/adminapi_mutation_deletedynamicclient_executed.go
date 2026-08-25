package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteDynamicClientExecuted event.Type = "admin_api.mutation.delete_dynamic_client.executed"
)

// AdminAPIMutationDeleteDynamicClientExecutedEventPayload carries the
// deleted client's identifying fields, read before the delete. Source is
// recorded because it determines what the deletion actually means: for DCR
// the client_id is permanently gone, while for CIMD the same client_id URL
// can produce a new record on its next resolution (docs/specs/dcr.md — New
// mutation).
type AdminAPIMutationDeleteDynamicClientExecutedEventPayload struct {
	ClientID   string                  `json:"client_id"`
	Source     model.OAuthClientSource `json:"source"`
	Kind       model.OAuthClientKind   `json:"kind"`
	ClientName string                  `json:"client_name,omitempty"`
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteDynamicClientExecuted
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

// FillContext sets ClientID for the same reason as
// OAuthClientRegisteredEventPayload: it makes "everything that ever happened
// to client X" a single indexed query on _audit_log.client_id rather than a
// jsonb scan over data->'payload'. This is a deliberate deviation from the
// other admin_api.mutation.* payloads, whose FillContext is empty — none of
// them is about an OAuth client, so none of them has a client id to put
// there.
func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) FillContext(ctx *event.Context) {
	ctx.ClientID = e.ClientID
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationDeleteDynamicClientExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteDynamicClientExecutedEventPayload{}
