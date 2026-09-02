package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OAuthClientRegistrationFailed event.Type = "oauth.client.registration.failed"
)

// OAuthClientRegistrationReason is the coarse, always-present bucket a
// registration failure falls into -- a small closed enum, unlike Message
// below.
type OAuthClientRegistrationReason string

const (
	// OAuthClientRegistrationReasonInvalidInitialAccessToken means the
	// Authorization header was malformed, an initial access token was
	// required and absent, or the token presented was unknown or expired.
	// Message distinguishes which: "malformed_header", "not_presented",
	// "unknown", "expired".
	OAuthClientRegistrationReasonInvalidInitialAccessToken OAuthClientRegistrationReason = "invalid_initial_access_token"
	// OAuthClientRegistrationReasonInvalidClientMetadata means the body
	// was not a JSON object, or failed a rule in dcr.md's Accepted Client
	// Metadata.
	OAuthClientRegistrationReasonInvalidClientMetadata OAuthClientRegistrationReason = "invalid_client_metadata"
	// OAuthClientRegistrationReasonLimitExceeded means the project is at
	// its oauth_client_dcr quota.
	OAuthClientRegistrationReasonLimitExceeded OAuthClientRegistrationReason = "limit_exceeded"
)

// OAuthClientRegistrationFailedEventPayload has no oracle constraint, unlike
// its CIMD counterpart: POST /oauth2/register fetches nothing, so there is
// no third party and no network reachability to leak, and it already
// returns a detailed RFC 7591 error to the caller. Message is therefore
// unrestricted -- withholding it from the project's own audit log would
// protect nothing while making the record useless.
type OAuthClientRegistrationFailedEventPayload struct {
	Reason OAuthClientRegistrationReason `json:"reason"`
	// Message names the specific cause within Reason. Free text, not a
	// fixed enum. Always set for "invalid_initial_access_token" and
	// "invalid_client_metadata"; empty for "limit_exceeded", which has no
	// sub-cases.
	Message string `json:"message,omitempty"`
	// UsageName and Quota are set only when Reason is "limit_exceeded".
	UsageName model.UsageName `json:"usage_name,omitempty"`
	Quota     int             `json:"quota,omitempty"`
	// InitialAccessToken identifies the token that was rejected. Present
	// only when Message is "expired" -- an unknown token has no row to
	// describe, and none was presented in the "not_presented" case. Shared
	// shape with oauth.client.registered, deliberately: see
	// EventPayloadInitialAccessToken's own doc comment.
	InitialAccessToken *EventPayloadInitialAccessToken `json:"initial_access_token,omitempty"`
}

func (e *OAuthClientRegistrationFailedEventPayload) NonBlockingEventType() event.Type {
	return OAuthClientRegistrationFailed
}

func (e *OAuthClientRegistrationFailedEventPayload) UserID() string { return "" }

func (e *OAuthClientRegistrationFailedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

// FillContext sets nothing: no client was created, so none exists to name.
// The record is still identifiable via event.Context's own
// IPAddress/UserAgent/Timestamp/AuditContext["http_url"], which is exactly
// what an admin needs to recognise someone working through a list of
// guessed initial access tokens.
func (e *OAuthClientRegistrationFailedEventPayload) FillContext(ctx *event.Context) {}

func (e *OAuthClientRegistrationFailedEventPayload) ForHook() bool  { return false }
func (e *OAuthClientRegistrationFailedEventPayload) ForAudit() bool { return true }

func (e *OAuthClientRegistrationFailedEventPayload) RequireReindexUserIDs() []string { return nil }
func (e *OAuthClientRegistrationFailedEventPayload) DeletedUserIDs() []string        { return nil }

var _ event.NonBlockingPayload = &OAuthClientRegistrationFailedEventPayload{}
