package handler

import (
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrInvalidRefreshToken = protocol.NewError("invalid_grant", "invalid refresh token")

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
}

func (s *TokenService) IssueOfflineGrant(
	client *config.OAuthClientConfig,
	opts IssueOfflineGrantOptions,
	resp protocol.TokenResponse,
) (offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {
	token := s.GenerateToken()
	tokenHash = oauth.HashToken(token)
	now := s.Clock.NowUTC()
	accessEvent := access.NewEvent(now, s.RemoteIP, s.UserAgentString)

	refreshToken := &oauth.OfflineGrantRefreshToken{
		TokenHash:       tokenHash,
		ClientID:        client.ClientID,
		CreatedAt:       now,
		Scopes:          opts.Scopes,
		AuthorizationID: opts.AuthorizationID,
	}

	offlineGrant = &oauth.OfflineGrant{
		AppID:        string(s.AppID),
		ID:           uuid.New(),
		IDPSessionID: opts.IDPSessionID,
		IdentityID:   opts.IdentityID,

		InitialClientID: client.ClientID,
		Scopes:          opts.Scopes,
		AuthorizationID: opts.AuthorizationID,
		TokenHash:       tokenHash,

		CreatedAt:       now,
		AuthenticatedAt: opts.AuthenticationInfo.AuthenticatedAt,

		Attrs: *session.NewAttrsFromAuthenticationInfo(opts.AuthenticationInfo),
		AccessInfo: access.Info{
			InitialAccess: accessEvent,
			LastAccess:    accessEvent,
		},

		DeviceInfo:              opts.DeviceInfo,
		SSOEnabled:              opts.SSOEnabled,
		App2AppDeviceKeyJWKJSON: "",

		RefreshTokens: []oauth.OfflineGrantRefreshToken{*refreshToken},
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

	err = s.OfflineGrants.CreateOfflineGrant(offlineGrant, expiry)
	if err != nil {
		return nil, "", err
	}

	err = s.AccessEvents.InitStream(offlineGrant.ID, &offlineGrant.AccessInfo.InitialAccess)
	if err != nil {
		return nil, "", err
	}

	if resp != nil {
		resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	}
	return offlineGrant, tokenHash, nil
}

func (s *TokenService) IssueAccessGrant(
	client *config.OAuthClientConfig,
	scopes []string,
	authzID string,
	userID string,
	sessionID string,
	sessionKind oauth.GrantSessionKind,
	refreshTokenHash string,
	resp protocol.TokenResponse,
) error {
	token := s.GenerateToken()
	now := s.Clock.NowUTC()

	accessGrant := &oauth.AccessGrant{
		AppID:            string(s.AppID),
		AuthorizationID:  authzID,
		SessionID:        sessionID,
		SessionKind:      sessionKind,
		CreatedAt:        now,
		ExpireAt:         now.Add(client.AccessTokenLifetime.Duration()),
		Scopes:           scopes,
		TokenHash:        oauth.HashToken(token),
		RefreshTokenHash: refreshTokenHash,
	}
	err := s.AccessGrants.CreateAccessGrant(accessGrant)
	if err != nil {
		return err
	}

	at, err := s.AccessTokenIssuer.EncodeAccessToken(client, accessGrant, userID, token)
	if err != nil {
		return err
	}

	resp.TokenType("Bearer")
	resp.AccessToken(at)
	resp.ExpiresIn(int(client.AccessTokenLifetime))
	return nil
}

func (s *TokenService) ParseRefreshToken(token string) (
	authz *oauth.Authorization, offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {
	token, grantID, err := oauth.DecodeRefreshToken(token)
	if err != nil {
		return nil, nil, "", ErrInvalidRefreshToken
	}

	offlineGrant, err = s.OfflineGrants.GetOfflineGrant(grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, nil, "", ErrInvalidRefreshToken
	} else if err != nil {
		return nil, nil, "", err
	}

	isValid, _, err := s.OfflineGrantService.IsValid(offlineGrant)
	if err != nil {
		return nil, nil, "", err
	}

	if !isValid {
		return nil, nil, "", ErrInvalidRefreshToken
	}

	tokenHash = oauth.HashToken(token)
	if !offlineGrant.MatchHash(tokenHash) {
		return nil, nil, "", ErrInvalidRefreshToken
	}

	offlineGrantSession, ok := offlineGrant.ToSession(tokenHash)
	if !ok {
		return nil, nil, "", ErrInvalidRefreshToken
	}

	authz, err = s.Authorizations.GetByID(offlineGrantSession.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, nil, "", ErrInvalidRefreshToken
	} else if err != nil {
		return nil, nil, "", err
	}

	// Standard session checking consider ErrUserNotFound and disabled as invalid.
	u, err := s.Users.GetRaw(offlineGrant.Attrs.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, nil, "", ErrInvalidRefreshToken
		}
		return nil, nil, "", err
	}
	err = u.AccountStatus().Check()
	if err != nil {
		return nil, nil, "", ErrInvalidRefreshToken
	}

	return authz, offlineGrant, tokenHash, nil
}
