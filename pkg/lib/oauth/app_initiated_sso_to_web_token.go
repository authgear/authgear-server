package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	AppInitiatedSSOToWebTokenLifetime = duration.Short
)

type AppInitiatedSSOToWebToken struct {
	AppID           string   `json:"app_id"`
	AuthorizationID string   `json:"authorization_id"`
	ClientID        string   `json:"client_id"`
	OfflineGrantID  string   `json:"offline_grant_id"`
	Scopes          []string `json:"scopes"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	TokenHash string    `json:"token_hash"`
}

type AppInitiatedSSOToWebTokenAccessGrantService interface {
	IssueAccessGrant(
		client *config.OAuthClientConfig,
		scopes []string,
		authzID string,
		userID string,
		sessionID string,
		sessionKind GrantSessionKind,
		refreshTokenHash string,
	) (*IssueAccessGrantResult, error)
}

type IssueAppInitiatedSSOToWebTokenResult struct {
	Token     string
	TokenHash string
	TokenType string
	ExpiresIn int
}

type IssueAppInitiatedSSOToWebTokenOptions struct {
	AppID           string
	ClientID        string
	OfflineGrantID  string
	AuthorizationID string
	Scopes          []string
}

func (s *AppInitiatedSSOToWebTokenService) IssueAppInitiatedSSOToWebToken(
	options *IssueAppInitiatedSSOToWebTokenOptions,
) (*IssueAppInitiatedSSOToWebTokenResult, error) {
	now := s.Clock.NowUTC()
	token := GenerateToken()
	tokenHash := HashToken(token)
	err := s.AppInitiatedSSOToWebTokens.CreateAppInitiatedSSOToWebToken(&AppInitiatedSSOToWebToken{
		AppID:           options.AppID,
		AuthorizationID: options.AuthorizationID,
		ClientID:        options.ClientID,
		OfflineGrantID:  options.OfflineGrantID,
		Scopes:          options.Scopes,

		CreatedAt: now,
		ExpireAt:  now.Add(AppInitiatedSSOToWebTokenLifetime),
		TokenHash: tokenHash,
	})
	if err != nil {
		return nil, err
	}

	return &IssueAppInitiatedSSOToWebTokenResult{
		Token:     token,
		TokenHash: tokenHash,
		TokenType: "Bearer",
		ExpiresIn: int(AppInitiatedSSOToWebTokenLifetime.Seconds()),
	}, nil
}

func (s *AppInitiatedSSOToWebTokenService) ExchangeForAccessToken(
	client *config.OAuthClientConfig,
	token string,
) (string, error) {
	tokenHash := HashToken(token)
	tokenModel, err := s.AppInitiatedSSOToWebTokens.GetAppInitiatedSSOToWebToken(tokenHash)
	if err != nil {
		return "", err
	}
	if tokenModel.ClientID != client.ClientID {
		return "", ErrUnmatchedClient
	}

	offlineGrant, err := s.OfflineGrants.GetOfflineGrant(tokenModel.OfflineGrantID)
	if err != nil {
		return "", err
	}

	newRefreshTokenResult, newOfflineGrant, err := s.OfflineGrantService.CreateNewRefreshToken(
		offlineGrant, tokenModel.ClientID, tokenModel.Scopes, tokenModel.AuthorizationID,
	)
	if err != nil {
		return "", err
	}
	offlineGrant = newOfflineGrant

	result, err := s.AccessGrantService.IssueAccessGrant(
		client,
		tokenModel.Scopes,
		tokenModel.AuthorizationID,
		offlineGrant.GetUserID(),
		offlineGrant.ID,
		GrantSessionKindOffline,
		newRefreshTokenResult.TokenHash,
	)

	if err != nil {
		return "", err
	}

	return result.Token, nil
}

type AppInitiatedSSOToWebTokenOfflineGrantService interface {
	CreateNewRefreshToken(
		grant *OfflineGrant,
		clientID string,
		scopes []string,
		authorizationID string,
	) (*CreateNewRefreshTokenResult, *OfflineGrant, error)
}
