package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OAuthClientResolutionFailed event.Type = "oauth.client.resolution.failed"
)

// OAuthClientResolutionReason is the coarse, always-present bucket a
// resolution failure falls into -- a small closed enum, unlike Message
// below. json:"reason" is deliberately the enum, not the free-text detail:
// the event type already says "failed", so what a reader needs next is a
// short, filterable category, not prose.
type OAuthClientResolutionReason string

const (
	// OAuthClientResolutionReasonUnavailable covers EVERY transport-level
	// failure as one indistinguishable bucket: DNS, blocked address,
	// connection refused, TLS failure, timeout, non-2xx, oversize,
	// unparseable body. The audit log is readable by the project admin, a
	// second and more privileged channel than the uniform HTTP error --
	// subdividing this would let a tenant admin probe the operator's
	// internal network via crafted client_ids. Never subdivide this, for
	// any reason -- Message is always empty here.
	OAuthClientResolutionReasonUnavailable OAuthClientResolutionReason = "unavailable"
	// OAuthClientResolutionReasonInvalid means a JSON object was retrieved
	// and failed a rule in cimd.md's Accepted Metadata Fields. Safe to
	// carry a Message (unlike Unavailable): reaching this reason proves a
	// parseable document was retrieved, so the message describes the
	// client author's own published content, not Authgear's network
	// reachability.
	OAuthClientResolutionReasonInvalid OAuthClientResolutionReason = "invalid"
	// OAuthClientResolutionReasonLimitExceeded means the document resolved
	// cleanly but the project is at its oauth_client_cimd quota, so no
	// record was created. Distinguishable per spec § Error Handling's
	// explicit carve-out from the uniform-error rule: it doesn't vary by
	// target host or reveal anything about network reachability.
	OAuthClientResolutionReasonLimitExceeded OAuthClientResolutionReason = "limit_exceeded"
)

type OAuthClientResolutionFailedEventPayload struct {
	ClientID string                      `json:"client_id"`
	Reason   OAuthClientResolutionReason `json:"reason"`
	// Message names the specific validation rule that failed, e.g.
	// "redirect_uris_missing". Free text, not a fixed enum -- new
	// validation rules add new values over time. Set only when Reason is
	// "invalid"; empty otherwise -- see the Unavailable/Invalid doc
	// comments above for why this split exists.
	Message string `json:"message,omitempty"`
	// UsageName and Quota are set only when Reason is "limit_exceeded".
	UsageName model.UsageName `json:"usage_name,omitempty"`
	Quota     int             `json:"quota,omitempty"`
	// ServedStaleRecord is true when a persisted record already existed and
	// was served despite this failure, false when the client was left
	// unresolvable. Always false when Reason is "limit_exceeded", which
	// only arises when there was no record to begin with. This is
	// Authgear's own state, not a property of the fetch target, so it is
	// safe to report -- and it is the difference between "your client is
	// degraded" and "your client is broken".
	ServedStaleRecord bool `json:"served_stale_record"`
}

func (e *OAuthClientResolutionFailedEventPayload) NonBlockingEventType() event.Type {
	return OAuthClientResolutionFailed
}

func (e *OAuthClientResolutionFailedEventPayload) UserID() string { return "" }

func (e *OAuthClientResolutionFailedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *OAuthClientResolutionFailedEventPayload) FillContext(ctx *event.Context) {
	ctx.ClientID = e.ClientID
}

func (e *OAuthClientResolutionFailedEventPayload) ForHook() bool  { return false }
func (e *OAuthClientResolutionFailedEventPayload) ForAudit() bool { return true }

func (e *OAuthClientResolutionFailedEventPayload) RequireReindexUserIDs() []string { return nil }
func (e *OAuthClientResolutionFailedEventPayload) DeletedUserIDs() []string        { return nil }

var _ event.NonBlockingPayload = &OAuthClientResolutionFailedEventPayload{}
