package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OAuthClientRegistered event.Type = "oauth.client.registered"
)

// OAuthClientRegisteredEventPayloadClient embeds
// model.OAuthClientRegisteredMetadata rather than repeating its fields --
// see that type's own comment for why: the same embed backs
// handler.RegistrationResponse, so the two cannot drift apart the way they
// once did (client_uri/logo_uri/tos_uri/policy_uri/
// token_endpoint_auth_method were present in the HTTP response but missing
// here).
type OAuthClientRegisteredEventPayloadClient struct {
	ClientID string                  `json:"client_id"`
	Source   model.OAuthClientSource `json:"source"`
	Kind     model.OAuthClientKind   `json:"kind"`
	model.OAuthClientRegisteredMetadata
}

type OAuthClientRegisteredEventPayload struct {
	Client OAuthClientRegisteredEventPayloadClient `json:"client"`

	// Request is the client metadata the caller asked for, as sent -- no
	// defaults applied, no normalization, nothing dropped. Client above is
	// what was actually registered, which can differ: an unimplemented
	// grant_type is silently dropped rather than refused, and
	// token_endpoint_auth_method is always ignored, so Request is what
	// makes either of those visible in the audit log at all. Shared type
	// with oauth.client.registration.failed's own Request field, since it
	// is the same record either way -- see
	// model.OAuthClientRegistrationRequest's own doc comment.
	Request model.OAuthClientRegistrationRequest `json:"request"`

	// InitialAccessToken is nil under open registration
	// (initial_access_token_required: false), and the `omitempty` drops the
	// key entirely — which is exactly the signal an auditor wants. An absent
	// key means "anyone who could reach the endpoint could have done this";
	// a present one names the IAT that authorized it, so a leaked IAT can be
	// traced to every client it registered. A pointer, not a value struct,
	// because "no IAT" must be distinguishable from "an IAT with empty
	// fields". docs/specs/event.md documents the key as absent, not null.
	//
	// model.EventPayloadInitialAccessToken is shared with
	// oauth.client.registration.failed's "expired" outcome, so a token
	// presents identically in both records.
	InitialAccessToken *model.EventPayloadInitialAccessToken `json:"initial_access_token,omitempty"`
}

func (e *OAuthClientRegisteredEventPayload) NonBlockingEventType() event.Type {
	return OAuthClientRegistered
}

func (e *OAuthClientRegisteredEventPayload) UserID() string { return "" }

func (e *OAuthClientRegisteredEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

// FillContext sets ClientID so the _audit_log.client_id column names the
// client that was just created, matching M2MTokenCreatedEventPayload
// (m2m_token_created.go), the only other event emitted from an OAuth
// endpoint with no authenticated user. docs/specs/event.md documents this as
// part of the event's contract ("context.client_id is the client_id of the
// newly registered client").
func (e *OAuthClientRegisteredEventPayload) FillContext(ctx *event.Context) {
	ctx.ClientID = e.Client.ClientID
}

func (e *OAuthClientRegisteredEventPayload) ForHook() bool  { return false }
func (e *OAuthClientRegisteredEventPayload) ForAudit() bool { return true }

func (e *OAuthClientRegisteredEventPayload) RequireReindexUserIDs() []string { return nil }
func (e *OAuthClientRegisteredEventPayload) DeletedUserIDs() []string        { return nil }

var _ event.NonBlockingPayload = &OAuthClientRegisteredEventPayload{}
