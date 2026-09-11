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
	ClientURI       string                  `json:"client_uri,omitempty"`
	LogoURI         string                  `json:"logo_uri,omitempty"`
	TOSURI          string                  `json:"tos_uri,omitempty"`
	PolicyURI       string                  `json:"policy_uri,omitempty"`
	ApplicationType string                  `json:"application_type,omitempty"`
	// TokenEndpointAuthMethod is always "none": CIMD ignores whatever a
	// document declares (see docs/specs/cimd.md's
	// token_endpoint_auth_method section) and always resolves the client
	// as public with no secret.
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

type OAuthClientResolvedEventPayload struct {
	// Client carries the post-resolution state in both cases, so a created
	// record is complete on its own and a changed record shows the
	// resulting client alongside its previous state.
	Client OAuthClientResolvedEventPayloadClient `json:"client"`
	// Document is the metadata document as fetched for this resolution --
	// no defaults applied, nothing dropped. Client above is what was
	// actually persisted, which can differ: an unimplemented grant_type is
	// silently dropped rather than refused, and token_endpoint_auth_method
	// is always ignored, so Document is what makes either of those visible
	// in the audit log at all. See model.OAuthClientResolutionDocument's
	// own doc comment.
	Document model.OAuthClientResolutionDocument `json:"document"`
	// Created is true on first resolution, false on a refetch that changed
	// something. An explicit discriminator: an auditor should not have to
	// infer it from OldClient being absent.
	Created bool `json:"created"`
	// OldClient is the client's state immediately before this resolution --
	// absent when Created is true (there is no "before"), and otherwise
	// always present: this event is never emitted for a refetch that
	// changed nothing (the routine hourly case, which carries no
	// information), so a present OldClient is guaranteed to differ from
	// Client in at least one field. Showing both states in full, rather
	// than a computed list of changed fields, is deliberate: the reader
	// does the diffing, which is simpler and cannot itself have a bug.
	OldClient *OAuthClientResolvedEventPayloadClient `json:"old_client,omitempty"`
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
