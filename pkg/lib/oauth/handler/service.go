package handler

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var errInvalidRefreshToken = protocol.NewError("invalid_grant", "invalid refresh token")

type TokenService struct {
	Request    *http.Request
	AppID      config.AppID
	Config     *config.OAuthConfig
	TrustProxy config.TrustProxy

	Authorizations    oauth.AuthorizationStore
	OfflineGrants     oauth.OfflineGrantStore
	AccessGrants      oauth.AccessGrantStore
	AccessEvents      *access.EventProvider
	AccessTokenIssuer AccessTokenIssuer
	GenerateToken     TokenGenerator
	Clock             clock.Clock
	Users             TokenHandlerUserFacade
}

func (s *TokenService) IssueOfflineGrant(
	client *config.OAuthClientConfig,
	opts IssueOfflineGrantOptions,
	resp protocol.TokenResponse,
) (*oauth.OfflineGrant, error) {
	token := s.GenerateToken()
	now := s.Clock.NowUTC()
	accessEvent := access.NewEvent(now, s.Request, bool(s.TrustProxy))

	offlineGrant := &oauth.OfflineGrant{
		AppID:           string(s.AppID),
		ID:              uuid.New(),
		AuthorizationID: opts.AuthorizationID,
		ClientID:        client.ClientID,
		IDPSessionID:    opts.IDPSessionID,
		IdentityID:      opts.IdentityID,

		CreatedAt:       now,
		AuthenticatedAt: opts.AuthenticationInfo.AuthenticatedAt,
		Scopes:          opts.Scopes,
		TokenHash:       oauth.HashToken(token),

		Attrs: *session.NewAttrsFromAuthenticationInfo(opts.AuthenticationInfo),
		AccessInfo: access.Info{
			InitialAccess: accessEvent,
			LastAccess:    accessEvent,
		},

		DeviceInfo: opts.DeviceInfo,
	}

	expiry := oauth.ComputeOfflineGrantExpiryWithClient(offlineGrant, client)
	err := s.OfflineGrants.CreateOfflineGrant(offlineGrant, expiry)
	if err != nil {
		return nil, err
	}

	err = s.AccessEvents.InitStream(offlineGrant.ID, &offlineGrant.AccessInfo.InitialAccess)
	if err != nil {
		return nil, err
	}

	resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	return offlineGrant, nil
}

func (s *TokenService) IssueAccessGrant(
	client *config.OAuthClientConfig,
	scopes []string,
	authzID string,
	userID string,
	sessionID string,
	sessionKind oauth.GrantSessionKind,
	resp protocol.TokenResponse,
) error {
	token := s.GenerateToken()
	now := s.Clock.NowUTC()

	accessGrant := &oauth.AccessGrant{
		AppID:           string(s.AppID),
		AuthorizationID: authzID,
		SessionID:       sessionID,
		SessionKind:     sessionKind,
		CreatedAt:       now,
		ExpireAt:        now.Add(client.AccessTokenLifetime.Duration()),
		Scopes:          scopes,
		TokenHash:       oauth.HashToken(token),
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

func (s *TokenService) ParseRefreshToken(token string) (*oauth.Authorization, *oauth.OfflineGrant, error) {
	token, grantID, err := oauth.DecodeRefreshToken(token)
	if err != nil {
		return nil, nil, errInvalidRefreshToken
	}

	offlineGrant, err := s.OfflineGrants.GetOfflineGrant(grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, nil, err
	}

	expiry, err := oauth.ComputeOfflineGrantExpiryWithClients(offlineGrant, s.Config)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, nil, err
	}

	if s.Clock.NowUTC().After(expiry) {
		return nil, nil, errInvalidRefreshToken
	}

	tokenHash := oauth.HashToken(token)
	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(offlineGrant.TokenHash)) != 1 {
		return nil, nil, errInvalidRefreshToken
	}

	authz, err := s.Authorizations.GetByID(offlineGrant.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, nil, err
	}

	// Standard session checking consider ErrUserNotFound and disabled as invalid.
	u, err := s.Users.GetRaw(offlineGrant.Attrs.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, nil, errInvalidRefreshToken
		}
		return nil, nil, err
	}
	err = u.CheckStatus()
	if err != nil {
		return nil, nil, errInvalidRefreshToken
	}

	return authz, offlineGrant, nil
}
