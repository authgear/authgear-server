package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

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

type PrepareUserAccessGrantByRefreshTokenOptions struct {
	oauth.PrepareUserAccessGrantOptions
	ShouldRotateRefreshToken bool
}

//go:generate go tool mockgen -source=service_token.go -destination=service_token_mock_test.go -package handler_test
type TokenServiceAuthorizationStore interface {
	oauth.AuthorizationStore
}

type TokenServiceOfflineGrantStore interface {
	oauth.OfflineGrantStore
}

type TokenServiceAccessGrantStore interface {
	oauth.AccessGrantStore
}

type TokenServiceOfflineGrantService interface {
	ComputeOfflineGrantExpiry(session *oauth.OfflineGrant) (expiry time.Time, err error)
	GetOfflineGrant(ctx context.Context, id string) (*oauth.OfflineGrant, error)
	CreateNewRefreshToken(
		ctx context.Context,
		options oauth.CreateNewRefreshTokenOptions,
	) (*oauth.CreateNewRefreshTokenResult, *oauth.OfflineGrant, error)
	RotateRefreshToken(
		ctx context.Context,
		options oauth.RotateRefreshTokenOptions,
	) (*oauth.RotateRefreshTokenResult, *oauth.OfflineGrant, error)
}

type TokenServiceAccessGrantService interface {
	PrepareUserAccessGrant(
		ctx context.Context,
		options oauth.PrepareUserAccessGrantOptions,
	) (oauth.PrepareUserAccessTokenResult, error)
}

type TokenServiceAccessTokenIssuer interface {
	EncodeClientAccessToken(ctx context.Context, options oauth.EncodeClientAccessTokenOptions) (string, error)
}

type TokenService struct {
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	AppID           config.AppID
	Config          *config.OAuthConfig

	Authorizations      TokenServiceAuthorizationStore
	OfflineGrants       TokenServiceOfflineGrantStore
	AccessGrants        TokenServiceAccessGrantStore
	OfflineGrantService TokenServiceOfflineGrantService
	AccessEvents        *access.EventProvider
	AccessTokenIssuer   TokenServiceAccessTokenIssuer
	GenerateToken       TokenGenerator
	Clock               clock.Clock
	Users               TokenHandlerUserFacade
	Events              EventService

	AccessGrantService TokenServiceAccessGrantService
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
		InitialTokenHash: tokenHash,
		ClientID:         client.ClientID,
		CreatedAt:        now,
		Scopes:           opts.Scopes,
		AuthorizationID:  opts.AuthorizationID,
		DPoPJKT:          opts.DPoPJKT,
		AccessInfo:       &accessInfo,
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

type PrepareUserAccessGrantByRefreshTokenResult struct {
	RotateRefreshTokenResult *oauth.RotateRefreshTokenResult
	PreparationResult        oauth.PrepareUserAccessTokenResult
}

func (s *TokenService) PrepareUserAccessGrantByRefreshToken(
	ctx context.Context,
	options PrepareUserAccessGrantByRefreshTokenOptions,
) (*PrepareUserAccessGrantByRefreshTokenResult, error) {
	result := &PrepareUserAccessGrantByRefreshTokenResult{}

	if options.ShouldRotateRefreshToken &&
		options.SessionLike.SessionType() == session.TypeOfflineGrant &&
		options.InitialRefreshTokenHash != "" {

		grant, err := s.OfflineGrantService.GetOfflineGrant(ctx, options.SessionLike.SessionID())
		if err != nil {
			return nil, err
		}

		rotateResult, _, err := s.OfflineGrantService.RotateRefreshToken(ctx,
			oauth.RotateRefreshTokenOptions{
				OfflineGrant:            grant,
				InitialRefreshTokenHash: options.InitialRefreshTokenHash,
			})
		if err != nil {
			return nil, err
		}
		result.RotateRefreshTokenResult = rotateResult
	}

	preparationResult, err := s.AccessGrantService.PrepareUserAccessGrant(
		ctx, options.PrepareUserAccessGrantOptions,
	)
	if err != nil {
		return nil, err
	}

	result.PreparationResult = preparationResult
	return result, nil
}

func (s *TokenService) ParseRefreshToken(ctx context.Context, token string) (
	authz *oauth.Authorization, offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {

	logger := TokenServiceLogger.GetLogger(ctx)

	dpopProof := dpop.GetDPoPProof(ctx)

	token, grantID, err := oauth.DecodeRefreshToken(token)
	if err != nil {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"failed to decode refresh token",
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	offlineGrant, err = s.OfflineGrantService.GetOfflineGrant(ctx, grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"failed to get offline grant: not found",
			slog.String("offline_grant_id", grantID),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	} else if err != nil {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"failed to get offline grant",
			slog.String("offline_grant_id", grantID),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", err
	}

	tokenHash = oauth.HashToken(token)
	if !offlineGrant.MatchCurrentHash(tokenHash) {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().Error(ctx,
			"failed to match refresh token hash",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	offlineGrantSession, ok := offlineGrant.ToSession(tokenHash)
	if !ok {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().Error(ctx,
			"failed to convert offline grant to session",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	}

	if dpopErr := offlineGrantSession.MatchDPoPJKT(dpopProof); dpopErr != nil {
		logger.WithSkipLogging().WithError(dpopErr).Error(ctx,
			fmt.Sprintf("failed to match dpop jkt on parse refresh token:%s", dpopErr.Message),
			slog.Bool("dpop_logs", true),
		)

		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().Error(ctx,
			"failed to match DPoP JKT",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", ErrInvalidDPoPKeyBinding
	}

	authz, err = s.Authorizations.GetByID(ctx, offlineGrantSession.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"failed to get authorization: not found",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", ErrInvalidRefreshToken
	} else if err != nil {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"failed to get authorization",
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", err
	}

	// Standard session checking consider ErrUserNotFound and disabled as invalid.
	u, err := s.Users.GetRaw(ctx, offlineGrant.Attrs.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			// NOTE(DEV-2982): This is for debugging the session lost problem
			logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
				"failed to get user: not found",
				slog.String("user_id", offlineGrant.GetUserID()),
				slog.String("offline_grant_id", offlineGrant.ID),
				slog.Bool("refresh_token_log", true),
			)
			return nil, nil, "", ErrInvalidRefreshToken
		}
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"failed to get user",
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.Bool("refresh_token_log", true),
		)
		return nil, nil, "", err
	}
	err = u.AccountStatus().Check()
	if err != nil {
		// NOTE(DEV-2982): This is for debugging the session lost problem
		logger.WithSkipLogging().WithSkipStackTrace().WithError(err).Error(ctx,
			"user account status check failed",
			slog.String("user_id", offlineGrant.GetUserID()),
			slog.String("offline_grant_id", offlineGrant.ID),
			slog.Bool("refresh_token_log", true),
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
