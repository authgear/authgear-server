package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OAuthClientRegistered event.Type = "oauth.client.registered"
)

type OAuthClientRegisteredEventPayloadClient struct {
	ClientID        string                  `json:"client_id"`
	Source          model.OAuthClientSource `json:"source"`
	Kind            model.OAuthClientKind   `json:"kind"`
	ClientName      string                  `json:"client_name,omitempty"`
	ApplicationType string                  `json:"application_type,omitempty"`
	RedirectURIs    []string                `json:"redirect_uris"`
	GrantTypes      []string                `json:"grant_types"`
	ResponseTypes   []string                `json:"response_types"`
}

type OAuthClientRegisteredEventPayload struct {
	Client OAuthClientRegisteredEventPayloadClient `json:"client"`

	// InitialAccessToken is nil under open registration
	// (initial_access_token_required: false), and the `omitempty` drops the
	// key entirely — which is exactly the signal an auditor wants. An absent
	// key means "anyone who could reach the endpoint could have done this";
	// a present one names the IAT that authorized it, so a leaked IAT can be
	// traced to every client it registered. A pointer, not a value struct,
	// because "no IAT" must be distinguishable from "an IAT with empty
	// fields". docs/specs/event.md documents the key as absent, not null.
	//
	// EventPayloadInitialAccessToken, not a type of its own: it is shared
	// with oauth.client.registration.failed's "expired" outcome, so a token
	// presents identically in both records.
	InitialAccessToken *EventPayloadInitialAccessToken `json:"initial_access_token,omitempty"`
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
