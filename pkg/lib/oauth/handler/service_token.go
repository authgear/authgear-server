package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

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

	AccessGrantService oauth.AccessGrantService
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
		DPoPJKT:         opts.DPoPJKT,
	}

	offlineGrant = &oauth.OfflineGrant{
		AppID:        string(s.AppID),
		ID:           uuid.New(),
		IDPSessionID: opts.IDPSessionID,
		IdentityID:   opts.IdentityID,

		InitialClientID: client.ClientID,

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

	if opts.IssueDeviceSecret {
		deviceSecretHash := s.IssueDeviceSecret(resp)
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

	err = s.OfflineGrants.CreateOfflineGrant(offlineGrant, expiry)
	if err != nil {
		return nil, "", err
	}

	err = s.AccessEvents.InitStream(offlineGrant.ID, expiry, &offlineGrant.AccessInfo.InitialAccess)
	if err != nil {
		return nil, "", err
	}

	if resp != nil {
		resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	}
	return offlineGrant, tokenHash, nil
}

func (s *TokenService) IssueRefreshTokenForOfflineGrant(
	offlineGrantID string,
	client *config.OAuthClientConfig,
	opts IssueOfflineGrantRefreshTokenOptions,
	resp protocol.TokenResponse,
) (offlineGrant *oauth.OfflineGrant, tokenHash string, err error) {
	offlineGrant, err = s.OfflineGrants.GetOfflineGrant(offlineGrantID)
	if err != nil {
		return nil, "", err
	}

	newRefreshTokenResult, newOfflineGrant, err := s.OfflineGrantService.CreateNewRefreshToken(
		offlineGrant, client.ClientID, opts.Scopes, opts.AuthorizationID, opts.DPoPJKT,
	)
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
	client *config.OAuthClientConfig,
	scopes []string,
	authzID string,
	userID string,
	sessionID string,
	sessionKind oauth.GrantSessionKind,
	refreshTokenHash string,
	resp protocol.TokenResponse,
) error {
	result, err := s.AccessGrantService.IssueAccessGrant(
		client, scopes, authzID, userID, sessionID, sessionKind, refreshTokenHash,
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

	dpopProof := dpop.GetDPoPProof(ctx)

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

	if !offlineGrantSession.MatchDPoPJKT(dpopProof) {
		return nil, nil, "", ErrInvalidDPoPKeyBinding
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

func (s *TokenService) IssueDeviceSecret(resp protocol.TokenResponse) (deviceSecretHash string) {
	deviceSecret := s.GenerateToken()
	deviceSecretHash = oauth.HashToken(deviceSecret)
	resp.DeviceSecret(deviceSecret)
	return deviceSecretHash
}
