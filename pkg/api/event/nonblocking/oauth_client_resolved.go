package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OAuthClientResolved event.Type = "oauth.client.resolved"
)

// OAuthClientResolvedEventPayloadClient is deliberately not shared with
// OAuthClientRegisteredEventPayloadClient (same field set minus the IAT):
// the two are independent wire contracts documented separately in
// event.md, and factoring them together would mean a future DCR-only field
// silently changes this event's shape too.
type OAuthClientResolvedEventPayloadClient struct {
	ClientID        string                  `json:"client_id"`
	Source          model.OAuthClientSource `json:"source"`
	Kind            model.OAuthClientKind   `json:"kind"`
	ClientName      string                  `json:"client_name,omitempty"`
	ApplicationType string                  `json:"application_type,omitempty"`
	RedirectURIs    []string                `json:"redirect_uris"`
	GrantTypes      []string                `json:"grant_types"`
	ResponseTypes   []string                `json:"response_types"`
}

// OAuthClientResolvedEventPayloadChange is one changed field. Old and New
// are both included: they came from a document the client published
// publicly, so there is nothing to redact, and "redirect_uris went from X
// to Y" is the point of the record. any, not separate string/slice change
// lists, because the fields being diffed are a mix of string and []string
// and this is a JSON document either way.
type OAuthClientResolvedEventPayloadChange struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
}

type OAuthClientResolvedEventPayload struct {
	// Client carries the post-resolution state in both cases, so a created
	// record is complete on its own and a changed record shows the
	// resulting client alongside the deltas.
	Client OAuthClientResolvedEventPayloadClient `json:"client"`
	// Created is true on first resolution, false on a refetch that changed
	// something. An explicit discriminator: an auditor should not have to
	// infer it from Changes being absent.
	Created bool `json:"created"`
	// Changes is absent when Created, and otherwise non-empty -- this event
	// is never emitted for a refetch that changed nothing (the routine
	// hourly case, which carries no information).
	Changes []OAuthClientResolvedEventPayloadChange `json:"changes,omitempty"`
}

func (e *OAuthClientResolvedEventPayload) NonBlockingEventType() event.Type {
	return OAuthClientResolved
}

func (e *OAuthClientResolvedEventPayload) UserID() string { return "" }

func (e *OAuthClientResolvedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

// FillContext sets ClientID, the only identifier this event carries --
// CIMD resolution happens at /oauth2/authorize before the user
// authenticates, so there is genuinely no user, exactly as for
// oauth.client.registered and m2m.token.created.
func (e *OAuthClientResolvedEventPayload) FillContext(ctx *event.Context) {
	ctx.ClientID = e.Client.ClientID
}

func (e *OAuthClientResolvedEventPayload) ForHook() bool  { return false }
func (e *OAuthClientResolvedEventPayload) ForAudit() bool { return true }

func (e *OAuthClientResolvedEventPayload) RequireReindexUserIDs() []string { return nil }
func (e *OAuthClientResolvedEventPayload) DeletedUserIDs() []string        { return nil }

var _ event.NonBlockingPayload = &OAuthClientResolvedEventPayload{}
