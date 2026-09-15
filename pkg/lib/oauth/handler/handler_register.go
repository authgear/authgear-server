package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dcr"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var RegistrationHandlerLogger = slogutil.NewLogger("oauth-dcr-register")

//go:generate go tool mockgen -source=handler_register.go -destination=handler_register_mock_test.go -package handler_test

type RegistrationHandlerDCRService interface {
	RegisterClient(ctx context.Context, options *dcr.RegisterClientOptions) (*model.OAuthClient, error)
	CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error)
	// LockForClientCount serializes concurrent registrations for this app so
	// the CountClientsBySource-then-RegisterClient sequence below is atomic
	// with respect to the configured oauth_client_dcr quota.
	LockForClientCount(ctx context.Context, source model.OAuthClientSource) error
}

type RegistrationHandlerIATService interface {
	ValidateAndGetByToken(ctx context.Context, plaintext string) (*model.OAuthInitialAccessToken, error)
}

type RegistrationHandlerRateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

type RegistrationHandlerUsageLimiter interface {
	CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error
	ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int)
}

type RegistrationHandler struct {
	Database    *appdb.Handle
	OAuthConfig *config.OAuthConfig
	DCR         RegistrationHandlerDCRService
	IAT         RegistrationHandlerIATService
	Clock       clock.Clock
	Events      EventService

	RemoteIP     httputil.RemoteIP
	RateLimiter  RegistrationHandlerRateLimiter
	UsageLimiter RegistrationHandlerUsageLimiter
}

// RegistrationResponse is the RFC 7591 §3.2.1 success response. There is
// deliberately no client_secret / client_secret_expires_at field: DCR
// clients are always public, per docs/specs/dcr.md.
//
// These fields are the registration as Authgear recorded it, not an echo
// of the request -- §3.2.1 requires all registered metadata, which is
// what makes substituting a requested value safe.
// RegistrationResponse embeds model.OAuthClientRegisteredMetadata rather
// than repeating its fields -- see that type's own comment for why: the
// same embed backs oauth.client.registered's audit event, so the two
// cannot drift apart the way they once did (client_uri/logo_uri/tos_uri/
// policy_uri/token_endpoint_auth_method were present here but missing from
// the event).
type RegistrationResponse struct {
	ClientID         string `json:"client_id"`
	ClientIDIssuedAt int64  `json:"client_id_issued_at"`
	model.OAuthClientRegisteredMetadata
}

// registrationRequestBody is the raw JSON shape of the POST /oauth2/register
// request body, decoded before being handed to dcr.ValidateAndNormalize.
type registrationRequestBody struct {
	ClientName      *string  `json:"client_name"`
	RedirectURIs    []string `json:"redirect_uris"`
	GrantTypes      []string `json:"grant_types"`
	ResponseTypes   []string `json:"response_types"`
	ApplicationType *string  `json:"application_type"`
	LogoURI         *string  `json:"logo_uri"`
	ClientURI       *string  `json:"client_uri"`
	TOSURI          *string  `json:"tos_uri"`
	PolicyURI       *string  `json:"policy_uri"`
	// TokenEndpointAuthMethod is decoded but never validated: it is kept
	// only so a failed registration can record what was asked for.
	TokenEndpointAuthMethod *string `json:"token_endpoint_auth_method"`
}

// auditRequest describes this body for the failure event, as sent. Only
// the failure paths with a decoded body call it; the rest pass nil.
func (b *registrationRequestBody) auditRequest() *model.OAuthClientRegistrationRequest {
	return &model.OAuthClientRegistrationRequest{
		ClientName:              derefStringOr(b.ClientName, ""),
		RedirectURIs:            b.RedirectURIs,
		GrantTypes:              b.GrantTypes,
		ResponseTypes:           b.ResponseTypes,
		ApplicationType:         derefStringOr(b.ApplicationType, ""),
		TokenEndpointAuthMethod: derefStringOr(b.TokenEndpointAuthMethod, ""),
		ClientURI:               derefStringOr(b.ClientURI, ""),
		LogoURI:                 derefStringOr(b.LogoURI, ""),
		TOSURI:                  derefStringOr(b.TOSURI, ""),
		PolicyURI:               derefStringOr(b.PolicyURI, ""),
	}
}

func (h *RegistrationHandler) checkRateLimit(ctx context.Context, spec ratelimit.BucketSpec) error {
	var err error

	failedReservation, allowErr := h.RateLimiter.Allow(ctx, spec)
	if allowErr != nil {
		err = allowErr
	} else if resvErr := failedReservation.Error(); resvErr != nil {
		err = resvErr
	}

	if err != nil && apierrors.IsKind(err, ratelimit.RateLimited) {
		return protocol.NewErrorStatusCode("x_rate_limited", "rate limit exceeded, please try again later.", http.StatusTooManyRequests)
	}
	return err
}

// Handle implements POST /oauth2/register per docs/specs/dcr.md's
// Registration Endpoint section.
func (h *RegistrationHandler) Handle(ctx context.Context, r *http.Request) (*RegistrationResponse, error) {
	if !h.OAuthConfig.DynamicClientRegistration.IsEnabled() {
		return nil, protocol.NewErrorStatusCode("access_denied", "dynamic client registration is not enabled", http.StatusForbidden)
	}

	// Both rate limits are consumed before the Authorization header is even
	// parsed, so an invalid IAT cannot be used to probe the endpoint more
	// cheaply. See docs/specs/dcr.md's Rate Limits section.
	rateLimits := h.OAuthConfig.DynamicClientRegistration.GetRateLimits()
	if err := h.checkRateLimit(ctx, NewBucketSpecOAuthRegisterPerIP(rateLimits, string(h.RemoteIP))); err != nil {
		return nil, err
	}
	if err := h.checkRateLimit(ctx, NewBucketSpecOAuthRegisterPerProject(rateLimits)); err != nil {
		return nil, err
	}

	token := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonInvalidInitialAccessToken, "malformed_header", "", 0, nil, nil)
			return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "invalid Authorization header", http.StatusUnauthorized)
		}
		token = strings.TrimPrefix(authHeader, prefix)
	}

	if token == "" && h.OAuthConfig.DynamicClientRegistration.IsInitialAccessTokenRequired() {
		h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonInvalidInitialAccessToken, "not_presented", "", 0, nil, nil)
		return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "an initial access token is required", http.StatusUnauthorized)
	}

	var body registrationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonInvalidClientMetadata, "malformed_json", "", 0, nil, nil)
		return nil, protocol.NewErrorStatusCode("invalid_client_metadata", "malformed JSON body", http.StatusBadRequest)
	}

	normalized, err := dcr.ValidateAndNormalize(&dcr.RegistrationRequest{
		ClientName:      body.ClientName,
		RedirectURIs:    body.RedirectURIs,
		GrantTypes:      body.GrantTypes,
		ResponseTypes:   body.ResponseTypes,
		ApplicationType: body.ApplicationType,
		LogoURI:         body.LogoURI,
		ClientURI:       body.ClientURI,
		TOSURI:          body.TOSURI,
		PolicyURI:       body.PolicyURI,
	})
	if err != nil {
		httpErr, message := mapDCRValidationError(err)
		h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonInvalidClientMetadata, message, "", 0, nil, body.auditRequest())
		return nil, httpErr
	}

	// The IAT lookup and the client insert must share one transaction: both
	// go through h.Database's SQLExecutor, which requires an active tx-like
	// context on every query.
	var client *model.OAuthClient
	var iat *model.OAuthInitialAccessToken
	var countBeforeCreate int
	err = h.Database.WithTx(ctx, func(ctx context.Context) error {
		c, i, count, err := h.registerClientInTx(ctx, token, normalized, body.auditRequest())
		iat = i // set even on error: see registerClientInTx's own comment
		if err != nil {
			return err
		}
		client = c
		countBeforeCreate = count
		return h.Events.DispatchEventOnCommit(ctx, newOAuthClientRegisteredEventPayload(client, iat, body.auditRequest()))
	})
	if err != nil {
		if errors.Is(err, dcr.ErrInitialAccessTokenNotFound) {
			h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonInvalidInitialAccessToken, "unknown", "", 0, nil, body.auditRequest())
			return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "invalid or expired initial access token", http.StatusUnauthorized)
		}
		if errors.Is(err, dcr.ErrInitialAccessTokenExpired) {
			h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonInvalidInitialAccessToken, "expired", "", 0, iat, body.auditRequest())
			return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "invalid or expired initial access token", http.StatusUnauthorized)
		}
		// Any other error -- including the limit_exceeded access_denied
		// error already dispatched above, and a pure infrastructure
		// failure, which is not an audit outcome -- is returned as-is.
		return nil, err
	}
	h.UsageLimiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientDCR, countBeforeCreate)

	return &RegistrationResponse{
		ClientID:                      client.ClientID,
		ClientIDIssuedAt:              client.CreatedAt.Unix(),
		OAuthClientRegisteredMetadata: model.NewOAuthClientRegisteredMetadata(client),
	}, nil
}

// registerClientInTx runs inside h.Database.WithTx: the IAT lookup, the
// quota check-then-insert, and the client insert all need the same
// transaction, and factoring them out of Handle keeps Handle's own
// cognitive complexity within the repo's lint budget.
//
// The returned *model.OAuthInitialAccessToken is non-nil whenever a token
// was presented, even when err is also non-nil -- ValidateAndGetByToken
// deliberately returns a non-nil token together with
// dcr.ErrInitialAccessTokenExpired, so the caller's audit event can
// describe the row even though registration is refused. Check err first;
// the token is for reporting only, never for authorizing.
func (h *RegistrationHandler) registerClientInTx(
	ctx context.Context,
	token string,
	normalized *dcr.NormalizedRegistration,
	auditRequest *model.OAuthClientRegistrationRequest,
) (client *model.OAuthClient, iat *model.OAuthInitialAccessToken, countBeforeCreate int, err error) {
	kind := model.OAuthClientKindThirdParty
	if token != "" {
		iat, err = h.IAT.ValidateAndGetByToken(ctx, token)
		if err != nil {
			return nil, iat, 0, err
		}
		if iat.Type == model.OAuthInitialAccessTokenTypeFirstParty {
			kind = model.OAuthClientKindFirstParty
		}
	}

	// Close the check-then-insert race between concurrent registrations for
	// the same app: a plain "SELECT COUNT(*) then INSERT" has a TOCTOU
	// window where two concurrent requests both observe a count under quota
	// and both proceed. Serialize per-app with a transaction-scoped
	// advisory lock.
	if err := h.DCR.LockForClientCount(ctx, model.OAuthClientSourceDCR); err != nil {
		return nil, iat, 0, err
	}

	clientCount, err := h.DCR.CountClientsBySource(ctx, model.OAuthClientSourceDCR)
	if err != nil {
		return nil, iat, 0, err
	}
	//nolint:gosec // G115
	count := int(clientCount)
	if limitErr := h.UsageLimiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, count); limitErr != nil {
		// Dispatched here, inside this same transaction, before returning:
		// DispatchEventImmediately writes through its own path rather than
		// this transaction, so the record survives the rollback this
		// triggers.
		usageName, quota, _ := usage.StandingUsageLimitDetails(limitErr)
		h.dispatchRegistrationFailed(ctx, nonblocking.OAuthClientRegistrationReasonLimitExceeded, "", usageName, quota, nil, auditRequest)
		return nil, iat, 0, protocol.NewErrorStatusCode("access_denied", "the project has reached its dynamic client registration limit", http.StatusForbidden)
	}

	client, err = h.DCR.RegisterClient(ctx, &dcr.RegisterClientOptions{
		Kind:         kind,
		Registration: normalized,
	})
	if err != nil {
		return nil, iat, 0, err
	}

	return client, iat, count, nil
}

// mapDCRValidationError maps a dcr.ValidateAndNormalize sentinel error to
// its exact (error, status) pair from docs/specs/dcr.md's Errors table, and
// to the audit-log message naming the rule that failed -- kept in this one
// function so a new validation rule cannot get an HTTP error without also
// getting an audit message. Only ErrDCRRedirectURIInvalid maps to
// invalid_redirect_uri; every other validation failure — including a
// missing redirect_uris, which the spec's causes table places under
// invalid_client_metadata rather than invalid_redirect_uri — maps to
// invalid_client_metadata. Unlike CIMD's uniform "unavailable" reason,
// there is no oracle constraint on message here: POST /oauth2/register
// fetches nothing, so there is no reachability to leak, and this function
// already puts the same detail in the HTTP response.
func mapDCRValidationError(err error) (httpErr error, message string) {
	switch {
	case errors.Is(err, dcr.ErrDCRRedirectURIsMissing):
		message = "redirect_uris_missing"
	case errors.Is(err, dcr.ErrDCRRedirectURIInvalid):
		return protocol.NewErrorStatusCode("invalid_redirect_uri", err.Error(), http.StatusBadRequest), "redirect_uri_invalid"
	case errors.Is(err, dcr.ErrDCRGrantTypeUnsupported):
		message = "grant_type_unsupported"
	case errors.Is(err, dcr.ErrDCRResponseTypeInconsistent):
		message = "response_type_inconsistent"
	case errors.Is(err, dcr.ErrDCRApplicationTypeUnsupported):
		message = "application_type_unsupported"
	case errors.Is(err, dcr.ErrDCRURIFieldNotHTTPS):
		message = "uri_field_not_https"
	default:
		// Unreachable in practice -- ValidateAndNormalize returns only the
		// sentinels above -- but never silently emit an empty message for a
		// genuinely new rule.
		message = "unknown"
	}
	return protocol.NewErrorStatusCode("invalid_client_metadata", err.Error(), http.StatusBadRequest), message
}

// dispatchRegistrationFailed builds and dispatches oauth.client.registration.failed.
// usageName/quota are set only when reason is limit_exceeded; iat is set
// only for the "expired" message; auditRequest is nil only where the
// request body was never decoded -- all mirroring
// OAuthClientRegistrationFailedEventPayload's own field comments.
func (h *RegistrationHandler) dispatchRegistrationFailed(
	ctx context.Context,
	reason nonblocking.OAuthClientRegistrationReason,
	message string,
	usageName model.UsageName,
	quota int,
	iat *model.OAuthInitialAccessToken,
	auditRequest *model.OAuthClientRegistrationRequest,
) {
	h.dispatchImmediately(ctx, &nonblocking.OAuthClientRegistrationFailedEventPayload{
		Reason:             reason,
		Message:            message,
		UsageName:          usageName,
		Quota:              quota,
		InitialAccessToken: model.NewEventPayloadInitialAccessToken(iat),
		Request:            auditRequest,
	})
}

// dispatchImmediately is used for oauth.client.registration.failed, which
// cannot use DispatchEventOnCommit: most failures precede any transaction,
// and the limit_exceeded path's transaction is about to roll back, so
// OnCommit would drop exactly the record that matters. Same IsInTx/ReadOnly
// branch as cimd.Service.dispatchImmediately and
// usage.Limiter.dispatchEventImmediately.
//
// The dispatch error is deliberately swallowed, not returned: an audit
// write failing must never turn a 400/401/403 into a 500.
func (h *RegistrationHandler) dispatchImmediately(ctx context.Context, payload event.NonBlockingPayload) {
	dispatch := func(ctx context.Context) error {
		return h.Events.DispatchEventImmediately(ctx, payload)
	}
	var err error
	if h.Database.IsInTx(ctx) {
		err = dispatch(ctx)
	} else {
		err = h.Database.ReadOnly(ctx, dispatch)
	}
	if err != nil {
		RegistrationHandlerLogger.GetLogger(ctx).WithError(err).
			Error(ctx, "dcr: failed to dispatch audit event")
	}
}

func derefStringOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// newOAuthClientRegisteredEventPayload builds the audit event payload for a
// successful registration. iat is nil under open registration
// (initial_access_token_required: false), in which case the payload's
// InitialAccessToken field is left nil too — see
// nonblocking.OAuthClientRegisteredEventPayload. auditRequest is the
// request body as sent, from registrationRequestBody.auditRequest() -- it
// is what lets the event show a grant_type or token_endpoint_auth_method
// the caller asked for that Client above silently dropped or ignored.
func newOAuthClientRegisteredEventPayload(client *model.OAuthClient, iat *model.OAuthInitialAccessToken, auditRequest *model.OAuthClientRegistrationRequest) *nonblocking.OAuthClientRegisteredEventPayload {
	payload := &nonblocking.OAuthClientRegisteredEventPayload{
		Client: nonblocking.OAuthClientRegisteredEventPayloadClient{
			ClientID:                      client.ClientID,
			Source:                        client.Source,
			Kind:                          client.Kind,
			OAuthClientRegisteredMetadata: model.NewOAuthClientRegisteredMetadata(client),
		},
		Request: *auditRequest,
	}
	payload.InitialAccessToken = model.NewEventPayloadInitialAccessToken(iat)
	return payload
}
