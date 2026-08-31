package cimd

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var ServiceLogger = slogutil.NewLogger("cimd-service")

// CIMDUnresolvable is returned for EVERY fetch and validation failure mode:
// blocked address, timeout, non-2xx, oversize, invalid JSON, client_id
// mismatch, failed field validation, and "not allowed by allowed_domains".
// One Kind, one reason, one message -- docs/specs/cimd.md § Authgear as an
// SSRF/Probing Oracle. The concrete cause is logged, never attached.
var CIMDUnresolvable = apierrors.Invalid.WithReason("CIMDUnresolvable")

// CIMDClientLimitExceeded is deliberately a DIFFERENT Kind. Spec § Error
// Handling: the client limit "is not a fetch-outcome signal -- it doesn't
// vary by target host or reveal anything about network reachability -- so
// it falls outside the uniform-error rule".
var CIMDClientLimitExceeded = apierrors.Forbidden.WithReason("CIMDClientLimitExceeded")

// ErrUnresolvable and ErrClientLimitExceeded use plain .New, never
// NewWithCause/NewWithInfo: Details is rendered into the JSON body, so
// attaching the underlying network error would defeat the whole design of
// a uniform, uninformative error. Log the cause separately (see below).
func ErrUnresolvable() error {
	return CIMDUnresolvable.New("client_id is not resolvable")
}

func ErrClientLimitExceeded() error {
	return CIMDClientLimitExceeded.New("the project has reached its CIMD client limit")
}

type ServiceOAuthClientCommands interface {
	UpsertCIMDClient(ctx context.Context, options *oauthclient.UpsertCIMDClientOptions) (*oauthclient.Client, bool, error)
	LockForClientCount(ctx context.Context, source oauthclient.Source) error
	CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error)
}

type ServiceOAuthClientQueries interface {
	GetClientByClientID(ctx context.Context, clientID string) (*oauthclient.Client, error)
}

type ServiceDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) error
	IsInTx(ctx context.Context) bool
	// ReadOnly is used only by dispatchImmediately, to give a dispatch
	// outside any transaction (a fetch failure, before step 8) a database
	// scope the same way usage.Limiter.dispatchEventImmediately does.
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

// ServiceRateLimiter is the seam for the CIMD rate-limits feature, bound to
// *RateLimiter (pkg/lib/cimd/ratelimit.go).
type ServiceRateLimiter interface {
	CheckFetchAllowed(ctx context.Context) error
}

// ServiceUsageLimiter is the seam for the CIMD client-limit feature, bound
// to *usage.Limiter.
type ServiceUsageLimiter interface {
	CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error
	ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int)
}

// ServiceEventService is the seam for CIMD's two audit events.
// DispatchEventOnCommit is used for oauth.client.resolved, which is only
// ever emitted from inside the upsert transaction; DispatchEventImmediately
// is used for oauth.client.resolution.failed, whose two paths either have
// no transaction at all or one that is about to roll back.
type ServiceEventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type Service struct {
	OAuthConfig *config.OAuthConfig
	// OAuthFeatureConfig supplies insecure_http_allowed. Already fanned out
	// by wire (pkg/lib/deps/deps_config.go).
	OAuthFeatureConfig *config.OAuthFeatureConfig
	Clock              clock.Clock
	Fetcher            *Fetcher
	Commands           ServiceOAuthClientCommands
	Queries            ServiceOAuthClientQueries
	Database           ServiceDatabase
	SingleFlight       *FetchSingleFlight
	RateLimiter        ServiceRateLimiter
	UsageLimiter       ServiceUsageLimiter
	Events             ServiceEventService
}

// EnsureClientResolved is the ONLY place in CIMD that performs an outbound
// network call, and it must be called from exactly one place --
// /oauth2/authorize, outside any database transaction (docs/specs/cimd.md §
// Where resolution happens). Any additional call site is a spec change.
//
// It returns nil for every client_id that is NOT a CIMD candidate -- a
// static client, a dcrc_ client, an unknown opaque string, or any URL when
// CIMD is disabled. "Not a candidate" is not an error: the caller proceeds
// to ordinary resolution. Only a client_id that IS a candidate and could
// not be resolved returns CIMDUnresolvable.
func (s *Service) EnsureClientResolved(ctx context.Context, clientID string) error {
	if s.Database.IsInTx(ctx) {
		// A 5-second outbound HTTP call inside a Postgres transaction would
		// pin a pooled connection for the duration; an attacker driving cold
		// client_ids at /oauth2/authorize would exhaust the pool long before
		// hitting any rate limit. This is a programming error, not a runtime
		// condition -- panic so the first test that gets it wrong catches it.
		panic("cimd: EnsureClientResolved must not be called inside a transaction")
	}

	cfg := s.OAuthConfig.ClientIDMetadataDocument
	if !cfg.IsEnabled() {
		return nil
	}

	// (1) Shape. No network access. Must use the same allowInsecureHTTP as
	// the read path's candidate check, or a client_id could be fetched here
	// and be unresolvable immediately afterwards.
	allowInsecureHTTP := s.OAuthFeatureConfig.GetClientIDMetadataDocument().IsInsecureHTTPAllowed()
	u, err := oauthclient.ParseCIMDClientID(clientID, allowInsecureHTTP)
	if err != nil {
		return nil
	}

	// (2) Static config wins, and is never fetched: spec § Client ID
	// Format's "pre-registering Client Identifier URLs" pattern. DCR needs
	// no equivalent check -- a dcrc_ id cannot parse as a URL.
	if _, ok := s.OAuthConfig.GetClient(clientID); ok {
		return nil
	}

	// (3) Freshness. One Redis GET in the common case. Deliberately BEFORE
	// the domain-trust check: allowed_domains must never affect a row that
	// already exists, fresh or stale (Part 1 §4.1 / D5a) -- only reordering
	// it this way lets a fresh row bypass the check entirely instead of
	// being incorrectly refused the moment its domain is delisted.
	existing, err := s.Queries.GetClientByClientID(ctx, clientID)
	switch {
	case err == nil:
		if existing.Source != model.OAuthClientSourceCIMD {
			return nil
		}
		if s.isFresh(existing) {
			return nil
		}
	case errors.Is(err, oauthclient.ErrDynamicClientNotFound):
		existing = nil
	default:
		// An infrastructure failure is not an unresolvable client.
		return err
	}

	// (4) Domain trust, before anything that could touch the network. Only
	// here -- not on the read path, and not for a row that already exists
	// (handled above) -- so removing a domain stops brand-new clients from
	// onboarding, and stops refetches (an existing-but-stale row on a now-
	// disallowed domain serves its last-known-good metadata forever,
	// exactly like a failed refetch would -- D5a's "freezes this one's
	// metadata permanently"), without ever touching a row that is still
	// fresh.
	if !oauthclient.IsCIMDClientIDAllowed(cfg.GetAllowedDomains(), u) {
		if existing != nil {
			return nil // serve the frozen stale record; never refetch it
		}
		return ErrUnresolvable()
	}

	// (5) Single-flight. A Redis failure degrades to a possible stampede,
	// which beats refusing every request.
	acquired, err := s.SingleFlight.Acquire(ctx, singleFlightPurposeDocument, clientID)
	if err != nil {
		acquired = true
	}
	if !acquired {
		if existing != nil {
			return nil
		}
		return ErrUnresolvable()
	}

	// (6) Rate limits. Consumed only when a fetch is actually about to
	// happen, so a popular fresh client never burns tokens.
	if err := s.RateLimiter.CheckFetchAllowed(ctx); err != nil {
		return err
	}

	// (7) The only network access in the feature.
	if allowInsecureHTTP && !strings.EqualFold(u.Scheme, "https") {
		ServiceLogger.GetLogger(ctx).Warn(ctx, "cimd: fetching a metadata document over plaintext http",
			slog.String("client_id", clientID),
			slog.String("flag", "oauth.client_id_metadata_document.insecure_http_allowed"))
	}
	body, fetchErr := s.Fetcher.Fetch(ctx, u)
	var doc *Document
	// The outcome is classified by WHICH function failed, never by
	// inspecting the error's content (D9): Fetcher.Fetch is always
	// "unavailable", ParseAndValidate is always "invalid". If a future
	// refactor let a transport error fall into the "invalid" branch, the
	// oracle-safety guarantee below would be gone.
	var validationErr error
	if fetchErr == nil {
		doc, validationErr = ParseAndValidate(clientID, body, allowInsecureHTTP)
	}
	if fetchErr != nil || validationErr != nil {
		cause := fetchErr
		outcome := nonblocking.OAuthClientResolutionOutcomeUnavailable
		reason := ""
		if fetchErr == nil {
			cause = validationErr
			outcome = nonblocking.OAuthClientResolutionOutcomeInvalid
			reason = documentErrorReason(validationErr)
		}
		// The cause is logged here and nowhere else; it never reaches a
		// response.
		ServiceLogger.GetLogger(ctx).WithError(cause).
			With(slog.String("client_id", clientID)).
			Info(ctx, "cimd: failed to resolve client metadata document")
		s.dispatchImmediately(ctx, &nonblocking.OAuthClientResolutionFailedEventPayload{
			ClientID:          clientID,
			Outcome:           outcome,
			Reason:            reason,
			ServedStaleRecord: existing != nil,
		})
		if existing != nil {
			return nil // serve the stale record
		}
		return ErrUnresolvable()
	}

	// (8) Persist. Short write transaction, opened only now.
	return s.Database.WithTx(ctx, func(ctx context.Context) error {
		return s.upsert(ctx, clientID, doc, existing)
	})
}

// documentErrorReason names the validation rule that failed, one per
// ErrDocument* sentinel. Safe to expose in the audit log's "invalid"
// outcome: reaching this function at all means a parseable JSON document
// was retrieved, so the reason describes the client author's own published
// content, not Authgear's network reachability (D7).
func documentErrorReason(err error) string {
	switch {
	case errors.Is(err, ErrDocumentNotJSONObject):
		return "not_json_object"
	case errors.Is(err, ErrDocumentClientIDMismatch):
		return "client_id_mismatch"
	case errors.Is(err, ErrDocumentRedirectURIsMissing):
		return "redirect_uris_missing"
	case errors.Is(err, ErrDocumentRedirectURIInvalid):
		return "redirect_uri_invalid"
	case errors.Is(err, ErrDocumentGrantTypeUnsupported):
		return "grant_type_unsupported"
	case errors.Is(err, ErrDocumentResponseTypeInconsistent):
		return "response_type_inconsistent"
	case errors.Is(err, ErrDocumentApplicationTypeUnsupported):
		return "application_type_unsupported"
	case errors.Is(err, ErrDocumentTokenEndpointAuthMethodNotAccepted):
		return "token_endpoint_auth_method_not_accepted"
	case errors.Is(err, ErrDocumentURIFieldNotHTTPS):
		return "uri_field_not_https"
	default:
		// Unreachable in practice -- ParseAndValidate returns only the
		// sentinels above -- but never silently emit an empty reason for a
		// genuinely new rule; that would look like a config problem was
		// swallowed rather than surfaced.
		return "unknown"
	}
}

// dispatchImmediately is used for oauth.client.resolution.failed, which
// cannot use DispatchEventOnCommit: on a fetch failure no transaction was
// ever opened, and on the limit_exceeded path the transaction is about to
// roll back, so OnCommit would drop exactly the record that matters. Same
// IsInTx/ReadOnly branch as usage.Limiter.dispatchEventImmediately.
//
// The dispatch error is deliberately swallowed, not returned: an audit
// write failing must never convert a resolvable client into an error, nor
// a clean failure into a server_error.
func (s *Service) dispatchImmediately(ctx context.Context, payload event.NonBlockingPayload) {
	dispatch := func(ctx context.Context) error {
		return s.Events.DispatchEventImmediately(ctx, payload)
	}
	var err error
	if s.Database.IsInTx(ctx) {
		err = dispatch(ctx)
	} else {
		err = s.Database.ReadOnly(ctx, dispatch)
	}
	if err != nil {
		ServiceLogger.GetLogger(ctx).WithError(err).
			Error(ctx, "cimd: failed to dispatch audit event")
	}
}

// isFresh reports false for a NULL LastFetchedAt: that row was written by
// something other than this service.
func (s *Service) isFresh(c *oauthclient.Client) bool {
	if c.LastFetchedAt == nil {
		return false
	}
	return s.Clock.NowUTC().Sub(*c.LastFetchedAt) < RefetchInterval
}

func (s *Service) upsert(ctx context.Context, clientID string, doc *Document, existing *oauthclient.Client) error {
	// Serialize concurrent first-resolutions for this app so the
	// count-then-create sequence is atomic with respect to the quota. Same
	// reasoning and the same helper POST /oauth2/register uses; the lock key
	// is already scoped per source, so a CIMD resolution never serializes
	// against a DCR registration.
	if err := s.Commands.LockForClientCount(ctx, model.OAuthClientSourceCIMD); err != nil {
		return err
	}

	clientCount, err := s.Commands.CountClientsBySource(ctx, model.OAuthClientSourceCIMD)
	if err != nil {
		return err
	}
	//nolint:gosec // G115 -- a client count cannot exceed MaxInt
	countBefore := int(clientCount)

	// Speculative, and re-decided below against `created`: whether this
	// resolution consumes a slot isn't knowable until the upsert runs. A
	// refetch of an existing client_id must succeed even at or over quota
	// (spec § Client Limit), while a brand-new one must be refused. Doing
	// the upsert first and rolling back on refusal makes both true with one
	// round trip -- a separate existence check first would itself be a
	// TOCTOU hazard even under the advisory lock above.
	limitErr := s.UsageLimiter.CheckStanding(ctx, model.UsageNameOAuthClientCIMD, countBefore)

	client, created, err := s.Commands.UpsertCIMDClient(ctx, &oauthclient.UpsertCIMDClientOptions{
		ClientID:        clientID,
		ApplicationType: doc.ApplicationType,
		ClientName:      doc.ClientName,
		ClientURI:       doc.ClientURI,
		LogoURI:         doc.LogoURI,
		TOSURI:          doc.TOSURI,
		PolicyURI:       doc.PolicyURI,
		RedirectURIs:    doc.RedirectURIs,
		GrantTypes:      doc.GrantTypes,
		ResponseTypes:   doc.ResponseTypes,
	})
	if err != nil {
		return err
	}

	if created && limitErr != nil {
		// Dispatched via Immediately, inside this same transaction, before
		// returning: DispatchEventImmediately writes through its own path
		// rather than this transaction, so the record survives the
		// rollback below. served_stale_record is always false here -- this
		// path only runs when there was no existing record to fall back to.
		usageName, quota, _ := usage.StandingUsageLimitDetails(limitErr)
		s.dispatchImmediately(ctx, &nonblocking.OAuthClientResolutionFailedEventPayload{
			ClientID:          clientID,
			Outcome:           nonblocking.OAuthClientResolutionOutcomeLimitExceeded,
			UsageName:         usageName,
			Quota:             quota,
			ServedStaleRecord: false,
		})
		// Rolls back the INSERT above (a non-nil error from inside
		// Database.WithTx rolls the transaction back). A distinct error
		// from the uniform CIMDUnresolvable one: spec § Error Handling
		// carves it out because it "doesn't vary by target host or reveal
		// anything about network reachability".
		return ErrClientLimitExceeded()
	}

	if created {
		// Fires alert/hook/event triggers for any quota threshold this
		// creation crossed. Only for a creation -- a refetch crosses none --
		// and only after the write is known to have happened.
		s.UsageLimiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientCIMD, countBefore)
	}

	// oauth.client.resolved: emitted on creation, or on a refetch that
	// actually changed something. Never for a refetch that changed
	// nothing -- the routine hourly case, which carries no information and
	// would otherwise bury the records that matter.
	var changes []nonblocking.OAuthClientResolvedEventPayloadChange
	if !created {
		changes = diffClientMetadata(existing, doc)
	}
	if created || len(changes) > 0 {
		payload := &nonblocking.OAuthClientResolvedEventPayload{
			Client:  oauthClientResolvedEventPayloadClient(client),
			Created: created,
			Changes: changes,
		}
		if err := s.Events.DispatchEventOnCommit(ctx, payload); err != nil {
			return err
		}
	}

	return nil
}

func oauthClientResolvedEventPayloadClient(c *oauthclient.Client) nonblocking.OAuthClientResolvedEventPayloadClient {
	return nonblocking.OAuthClientResolvedEventPayloadClient{
		ClientID:        c.ClientID,
		Source:          c.Source,
		Kind:            c.Kind,
		ClientName:      derefStringOr(c.ClientName, ""),
		ApplicationType: c.ApplicationType,
		RedirectURIs:    c.RedirectURIs,
		GrantTypes:      c.GrantTypes,
		ResponseTypes:   c.ResponseTypes,
	}
}

func derefStringOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// diffClientMetadata compares existing (already in hand from the freshness
// check) against doc (the freshly fetched document) and reports every
// document-derived field that changed. Never last_fetched_at/updated_at,
// which change on every refetch by construction -- including them would
// make the event fire hourly for every client and render it useless.
//
// existing may be up to 5 minutes stale (the freshness read is cache-first).
// The only writer to this row is this same path under a single-flight
// guard, so a false "changed" is very unlikely -- and if it happens the
// record is a duplicate, not a wrong one, which is the acceptable
// direction. This does not re-read the row to close that gap; that would
// put a Postgres round trip on the hot path to improve an audit record.
func diffClientMetadata(existing *oauthclient.Client, doc *Document) []nonblocking.OAuthClientResolvedEventPayloadChange {
	var changes []nonblocking.OAuthClientResolvedEventPayloadChange

	addString := func(field string, old *string, new *string) {
		// Normalize the same way Store.NewClient does: an explicit "" is
		// stored as nil, so a document that switches between omitting a
		// field and sending "" must not register as a change.
		if old != nil && *old == "" {
			old = nil
		}
		if new != nil && *new == "" {
			new = nil
		}
		oldVal, newVal := derefStringOr(old, ""), derefStringOr(new, "")
		if oldVal == newVal {
			return
		}
		changes = append(changes, nonblocking.OAuthClientResolvedEventPayloadChange{
			Field: field,
			Old:   stringOrNil(old),
			New:   stringOrNil(new),
		})
	}

	addSet := func(field string, old []string, new []string) {
		// Order carries no meaning in these fields, so a document that
		// merely reorders entries is not a change worth a record. The
		// reported New value is the value as stored, not a sorted set.
		if setsEqual(old, new) {
			return
		}
		changes = append(changes, nonblocking.OAuthClientResolvedEventPayloadChange{
			Field: field,
			Old:   old,
			New:   new,
		})
	}

	addString("client_name", existing.ClientName, doc.ClientName)
	addString("client_uri", existing.ClientURI, doc.ClientURI)
	addString("logo_uri", existing.LogoURI, doc.LogoURI)
	addString("tos_uri", existing.TOSURI, doc.TOSURI)
	addString("policy_uri", existing.PolicyURI, doc.PolicyURI)
	if existing.ApplicationType != doc.ApplicationType {
		changes = append(changes, nonblocking.OAuthClientResolvedEventPayloadChange{
			Field: "application_type",
			Old:   existing.ApplicationType,
			New:   doc.ApplicationType,
		})
	}
	addSet("redirect_uris", existing.RedirectURIs, doc.RedirectURIs)
	addSet("grant_types", existing.GrantTypes, doc.GrantTypes)
	addSet("response_types", existing.ResponseTypes, doc.ResponseTypes)

	return changes
}

func stringOrNil(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func setsEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	setA := setutil.NewSetFromSlice(a, setutil.Identity)
	for _, v := range b {
		if !setA.Has(v) {
			return false
		}
	}
	return true
}
