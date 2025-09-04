package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var TokenServiceLogger = slogutil.NewLogger("oauth-token-service")

var ErrInvalidRefreshToken = protocol.NewError("invalid_grant", "invalid refresh token")
var ErrInvalidDPoPKeyBinding = protocol.NewError(dpop.InvalidDPoPProof, "Invalid DPoP key binding")

type IssueOfflineGrantOptions struct {
	AuthenticationInfo authenticationinfo.T
	Scopes             []string
	AuthorizationID    string
	IDPSessionID       string
	DeviceInfo         map[string]interface{}
	IdentityID         string
	SSOEnabled         bool
	App2AppDeviceKey   jwk.Key
	IssueDeviceSecret  bool
	DPoPJKT            string
}

type IssueOfflineGrantRefreshTokenOptions struct {
	Scopes          []string
	AuthorizationID string
	DPoPJKT         string
}

type ClientCredentialsAccessTokenOptions struct {
	ResourceURI        string
	Scopes             []string
	ClientConfig       *config.OAuthClientConfig
	MaskedClientSecret string
	Resource           *resourcescope.Resource
}

type TokenService struct {
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	AppID           config.AppID
	Config          *config.OAuthConfig

	Authorizations      oauth.AuthorizationStore
	OfflineGrants       oauth.OfflineGrantStore
	AccessGrants        oauth.AccessGrantStore
	OfflineGrantService oauth.OfflineGrantService
	AccessEvents        *access.EventProvider
	AccessTokenIssuer   AccessTokenIssuer
	GenerateToken       TokenGenerator
	Clock               clock.Clock
	Users               TokenHandlerUserFacade
	Events              EventService

	AccessGrantService oauth.AccessGrantService
}

func (s *TokenService) IssueOfflineGrant(
	ctx context.Context,
	client *config.OAuthClientConfig,
	opts IssueOfflineGrantOptions,
	resp protocol.TokenResponse,
) (offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {
	token := s.GenerateToken()
	tokenHash = oauth.HashToken(token)
	now := s.Clock.NowUTC()
	accessEvent := access.NewEvent(now, s.RemoteIP, s.UserAgentString)

	accessInfo := access.Info{
		InitialAccess: accessEvent,
		LastAccess:    accessEvent,
	}

	refreshToken := &oauth.OfflineGrantRefreshToken{
		TokenHash:       tokenHash,
		ClientID:        client.ClientID,
		CreatedAt:       now,
		Scopes:          opts.Scopes,
		AuthorizationID: opts.AuthorizationID,
		DPoPJKT:         opts.DPoPJKT,
		AccessInfo:      &accessInfo,
	}

	offlineGrant = &oauth.OfflineGrant{
		AppID:        string(s.AppID),
		ID:           uuid.New(),
		IDPSessionID: opts.IDPSessionID,
		IdentityID:   opts.IdentityID,

		InitialClientID: client.ClientID,

		CreatedAt:       now,
		AuthenticatedAt: opts.AuthenticationInfo.AuthenticatedAt,

		Attrs:      *session.NewAttrsFromAuthenticationInfo(opts.AuthenticationInfo),
		AccessInfo: accessInfo,

		DeviceInfo:              opts.DeviceInfo,
		SSOEnabled:              opts.SSOEnabled,
		App2AppDeviceKeyJWKJSON: "",

		RefreshTokens: []oauth.OfflineGrantRefreshToken{*refreshToken},
	}

	if opts.IssueDeviceSecret {
		deviceSecretHash := s.IssueDeviceSecret(ctx, resp)
		offlineGrant.DeviceSecretHash = deviceSecretHash
		offlineGrant.DeviceSecretDPoPJKT = opts.DPoPJKT
	}

	if opts.App2AppDeviceKey != nil {
		keyStr, err := json.Marshal(opts.App2AppDeviceKey)
		if err != nil {
			return nil, "", err
		}
		offlineGrant.App2AppDeviceKeyJWKJSON = string(keyStr)
	}

	expiry, err := s.OfflineGrantService.ComputeOfflineGrantExpiry(offlineGrant)
	if err != nil {
		return nil, "", err
	}
	offlineGrant.ExpireAtForResolvedSession = expiry

	err = s.OfflineGrants.CreateOfflineGrant(ctx, offlineGrant)
	if err != nil {
		return nil, "", err
	}

	err = s.AccessEvents.InitStream(ctx, offlineGrant.ID, expiry, &offlineGrant.AccessInfo.InitialAccess)
	if err != nil {
		return nil, "", err
	}

	if resp != nil {
		resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	}
	return offlineGrant, tokenHash, nil
}

func (s *TokenService) IssueRefreshTokenForOfflineGrant(
	ctx context.Context,
	offlineGrantID string,
	client *config.OAuthClientConfig,
	opts IssueOfflineGrantRefreshTokenOptions,
	resp protocol.TokenResponse,
) (offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {
	offlineGrant, err = s.OfflineGrantService.GetOfflineGrant(ctx, offlineGrantID)
	if err != nil {
		return nil, "", err
	}

	newRefreshTokenResult, newOfflineGrant, err := s.OfflineGrantService.CreateNewRefreshToken(ctx, oauth.CreateNewRefreshTokenOptions{
		OfflineGrant:    offlineGrant,
		ClientID:        client.ClientID,
		Scopes:          opts.Scopes,
		AuthorizationID: opts.AuthorizationID,
		DPoPJKT:         opts.DPoPJKT,
	})
	if err != nil {
		return nil, "", err
	}
	offlineGrant = newOfflineGrant

	if resp != nil {
		resp.RefreshToken(oauth.EncodeRefreshToken(newRefreshTokenResult.Token, offlineGrant.ID))
	}

	return newOfflineGrant, newRefreshTokenResult.TokenHash, nil
}

func (s *TokenService) IssueAccessGrant(
	ctx context.Context,
	options oauth.IssueAccessGrantOptions,
	resp protocol.TokenResponse,
) error {
	result, err := s.AccessGrantService.IssueAccessGrant(
		ctx, options,
	)
	if err != nil {
		return err
	}

	resp.TokenType(result.TokenType)
	resp.AccessToken(result.Token)
	resp.ExpiresIn(result.ExpiresIn)
	return nil
}

func (s *TokenService) ParseRefreshToken(ctx context.Context, token string) (
	authz *oauth.Authorization, offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {

	logger := TokenServiceLogger.GetLogger(ctx)

	dpopProof := dpop.GetDPoPProof(ctx)

	token, grantID, err := oauth.DecodeRefreshToken(token)
	if err != nil {
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"failed to decode refresh token",
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	offlineGrant, err = s.OfflineGrantService.GetOfflineGrant(ctx, grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"failed to get offline grant: not found",
			slog.String("offline_grant_id", grantID),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	} else if err != nil {
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"failed to get offline grant",
			slog.String("offline_grant_id", grantID),
		)
		return nil, nil, "", err
	}

	tokenHash = oauth.HashToken(token)
	if !offlineGrant.MatchHash(tokenHash) {
		logger.WithSkipLogging().Error(ctx,
			"failed to match refresh token hash",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	offlineGrantSession, ok := offlineGrant.ToSession(tokenHash)
	if !ok {
		logger.WithSkipLogging().Error(ctx,
			"failed to convert offline grant to session",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	if !offlineGrantSession.MatchDPoPJKT(dpopProof) {
		logger.WithSkipLogging().Error(ctx,
			"failed to match DPoP JKT",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
		)
		return nil, nil, "", ErrInvalidDPoPKeyBinding
	}

	authz, err = s.Authorizations.GetByID(ctx, offlineGrantSession.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"failed to get authorization: not found",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	} else if err != nil {
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"failed to get authorization",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
		)
		return nil, nil, "", err
	}

	// Standard session checking consider ErrUserNotFound and disabled as invalid.
	u, err := s.Users.GetRaw(ctx, offlineGrant.Attrs.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.WithSkipLogging().WithError(err).Error(ctx,
				"failed to get user: not found",
				slog.String("user_id", offlineGrant.GetUserID()),
				slog.String("offline_grant_id", offlineGrant.ID),
			)
			return nil, nil, "", ErrInvalidRefreshToken
		}
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"failed to get user",
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.String("offline_grant_id", offlineGrant.ID),
		)
		return nil, nil, "", err
	}
	err = u.AccountStatus().Check()
	if err != nil {
		logger.WithSkipLogging().WithError(err).Error(ctx,
			"user account status check failed",
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.String("offline_grant_id", offlineGrant.ID),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	return authz, offlineGrant, tokenHash, nil
}

func (s *TokenService) IssueDeviceSecret(ctx context.Context, resp protocol.TokenResponse) (deviceSecretHash string) {
	deviceSecret := s.GenerateToken()
	deviceSecretHash = oauth.HashToken(deviceSecret)
	resp.DeviceSecret(deviceSecret)
	return deviceSecretHash
}

func (s *TokenService) IssueClientCredentialsAccessToken(ctx context.Context, options ClientCredentialsAccessTokenOptions, resp protocol.TokenResponse) error {
	token := s.GenerateToken()
	now := s.Clock.NowUTC()
	expireAt := now.Add(options.ClientConfig.AccessTokenLifetime.Duration())

	scope := strings.Join(options.Scopes, " ")
	encodedToken, err := s.AccessTokenIssuer.EncodeClientAccessToken(ctx, oauth.EncodeClientAccessTokenOptions{
		OriginalToken: token,
		ClientConfig:  options.ClientConfig,
		ResourceURI:   options.ResourceURI,
		Scope:         scope,
		CreatedAt:     now,
		ExpireAt:      expireAt,
	})
	if err != nil {
		return err
	}

	resp.TokenType("Bearer")
	resp.AccessToken(encodedToken)
	resp.ExpiresIn(int(options.ClientConfig.AccessTokenLifetime.Duration().Seconds()))
	resp.Scope(scope)

	err = s.Events.DispatchEventOnCommit(ctx, &nonblocking.M2MTokenCreatedEventPayload{
		ClientID:     options.ClientConfig.ClientID,
		ClientSecret: options.MaskedClientSecret,
	})
	if err != nil {
		return err
	}

	return nil
}
