package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dcr"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RegistrationHandlerDCRService interface {
	RegisterClient(ctx context.Context, options *dcr.RegisterClientOptions) (*model.OAuthClient, error)
	// CountClientsBySource is present now for symmetry; wired up by the
	// client usage-limit check (docs/plans/dcr/2026-08-17-05-client-usage-limit.md).
	CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error)
}

type RegistrationHandlerIATService interface {
	ValidateAndGetByToken(ctx context.Context, plaintext string) (*model.OAuthInitialAccessToken, error)
}

type RegistrationHandlerRateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

type RegistrationHandler struct {
	Database    *appdb.Handle
	OAuthConfig *config.OAuthConfig
	DCR         RegistrationHandlerDCRService
	IAT         RegistrationHandlerIATService
	Clock       clock.Clock

	RemoteIP    httputil.RemoteIP
	RateLimiter RegistrationHandlerRateLimiter
}

// RegistrationResponse is the RFC 7591 §3.2.1 success response. There is
// deliberately no client_secret / client_secret_expires_at field: DCR
// clients are always public, per docs/specs/dcr.md.
type RegistrationResponse struct {
	ClientID         string   `json:"client_id"`
	ClientIDIssuedAt int64    `json:"client_id_issued_at"`
	ClientName       string   `json:"client_name,omitempty"`
	RedirectURIs     []string `json:"redirect_uris"`
	GrantTypes       []string `json:"grant_types"`
	ResponseTypes    []string `json:"response_types"`
	ApplicationType  string   `json:"application_type"`
	ClientURI        string   `json:"client_uri,omitempty"`
	LogoURI          string   `json:"logo_uri,omitempty"`
	TOSURI           string   `json:"tos_uri,omitempty"`
	PolicyURI        string   `json:"policy_uri,omitempty"`
}

// registrationRequestBody is the raw JSON shape of the POST /oauth2/register
// request body, decoded before being handed to dcr.ValidateAndNormalize.
type registrationRequestBody struct {
	ClientName              *string  `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	ApplicationType         *string  `json:"application_type"`
	LogoURI                 *string  `json:"logo_uri"`
	ClientURI               *string  `json:"client_uri"`
	TOSURI                  *string  `json:"tos_uri"`
	PolicyURI               *string  `json:"policy_uri"`
	TokenEndpointAuthMethod *string  `json:"token_endpoint_auth_method"`
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
			return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "invalid Authorization header", http.StatusUnauthorized)
		}
		token = strings.TrimPrefix(authHeader, prefix)
	}

	var kind oauthclient.Kind
	if token == "" {
		if h.OAuthConfig.DynamicClientRegistration.IsInitialAccessTokenRequired() {
			return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "an initial access token is required", http.StatusUnauthorized)
		}
		kind = model.OAuthClientKindThirdParty
	} else {
		iat, err := h.IAT.ValidateAndGetByToken(ctx, token)
		if err != nil {
			if errors.Is(err, dcr.ErrInitialAccessTokenNotFound) {
				return nil, protocol.NewErrorStatusCode("invalid_initial_access_token", "invalid or expired initial access token", http.StatusUnauthorized)
			}
			return nil, err
		}
		if iat.Type == model.OAuthInitialAccessTokenTypeFirstParty {
			kind = model.OAuthClientKindFirstParty
		} else {
			kind = model.OAuthClientKindThirdParty
		}
	}

	var body registrationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, protocol.NewErrorStatusCode("invalid_client_metadata", "malformed JSON body", http.StatusBadRequest)
	}

	normalized, err := dcr.ValidateAndNormalize(&dcr.RegistrationRequest{
		ClientName:              body.ClientName,
		RedirectURIs:            body.RedirectURIs,
		GrantTypes:              body.GrantTypes,
		ResponseTypes:           body.ResponseTypes,
		ApplicationType:         body.ApplicationType,
		LogoURI:                 body.LogoURI,
		ClientURI:               body.ClientURI,
		TOSURI:                  body.TOSURI,
		PolicyURI:               body.PolicyURI,
		TokenEndpointAuthMethod: body.TokenEndpointAuthMethod,
	})
	if err != nil {
		return nil, mapDCRValidationError(err)
	}

	var client *model.OAuthClient
	err = h.Database.WithTx(ctx, func(ctx context.Context) error {
		c, err := h.DCR.RegisterClient(ctx, &dcr.RegisterClientOptions{
			Kind:         kind,
			Registration: normalized,
		})
		if err != nil {
			return err
		}
		client = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RegistrationResponse{
		ClientID:         client.ClientID,
		ClientIDIssuedAt: client.CreatedAt.Unix(),
		ClientName:       client.Name,
		RedirectURIs:     client.RedirectURIs,
		GrantTypes:       client.GrantTypes,
		ResponseTypes:    client.ResponseTypes,
		ApplicationType:  derefStringOr(client.ApplicationType, ""),
		ClientURI:        derefStringOr(client.ClientURI, ""),
		LogoURI:          derefStringOr(client.LogoURI, ""),
		TOSURI:           derefStringOr(client.TOSURI, ""),
		PolicyURI:        derefStringOr(client.PolicyURI, ""),
	}, nil
}

// mapDCRValidationError maps a dcr.ValidateAndNormalize sentinel error to
// its exact (error, status) pair from docs/specs/dcr.md's Errors table.
// Only ErrDCRRedirectURIInvalid maps to invalid_redirect_uri; every other
// validation failure — including a missing redirect_uris, which the
// spec's causes table places under invalid_client_metadata rather than
// invalid_redirect_uri — maps to invalid_client_metadata.
func mapDCRValidationError(err error) error {
	if errors.Is(err, dcr.ErrDCRRedirectURIInvalid) {
		return protocol.NewErrorStatusCode("invalid_redirect_uri", err.Error(), http.StatusBadRequest)
	}
	return protocol.NewErrorStatusCode("invalid_client_metadata", err.Error(), http.StatusBadRequest)
}

func derefStringOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}
