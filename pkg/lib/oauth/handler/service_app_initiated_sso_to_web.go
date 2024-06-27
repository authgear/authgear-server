package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AppInitiatedSSOToWebTokenServiceImpl struct {
	Clock clock.Clock

	AppInitiatedSSOToWebTokens oauth.AppInitiatedSSOToWebTokenStore
	OfflineGrants              oauth.OfflineGrantStore
	AccessGrantService         oauth.AppInitiatedSSOToWebTokenAccessGrantService
	OfflineGrantService        oauth.AppInitiatedSSOToWebTokenOfflineGrantService
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

func (s *AppInitiatedSSOToWebTokenServiceImpl) IssueAppInitiatedSSOToWebToken(
	options *IssueAppInitiatedSSOToWebTokenOptions,
) (*IssueAppInitiatedSSOToWebTokenResult, error) {
	now := s.Clock.NowUTC()
	token := oauth.GenerateToken()
	tokenHash := oauth.HashToken(token)
	err := s.AppInitiatedSSOToWebTokens.CreateAppInitiatedSSOToWebToken(&oauth.AppInitiatedSSOToWebToken{
		AppID:           options.AppID,
		AuthorizationID: options.AuthorizationID,
		ClientID:        options.ClientID,
		OfflineGrantID:  options.OfflineGrantID,
		Scopes:          options.Scopes,

		CreatedAt: now,
		ExpireAt:  now.Add(oauth.AppInitiatedSSOToWebTokenLifetime),
		TokenHash: tokenHash,
	})
	if err != nil {
		return nil, err
	}

	return &IssueAppInitiatedSSOToWebTokenResult{
		Token:     token,
		TokenHash: tokenHash,
		TokenType: "Bearer",
		ExpiresIn: int(oauth.AppInitiatedSSOToWebTokenLifetime.Seconds()),
	}, nil
}

func (s *AppInitiatedSSOToWebTokenServiceImpl) ExchangeForAccessToken(
	client *config.OAuthClientConfig,
	sessionID string,
	token string,
) (string, error) {
	tokenHash := oauth.HashToken(token)
	tokenModel, err := s.AppInitiatedSSOToWebTokens.GetAppInitiatedSSOToWebToken(tokenHash)
	if err != nil {
		return "", err
	}
	if tokenModel.ClientID != client.ClientID {
		return "", oauth.ErrUnmatchedClient
	}
	if tokenModel.OfflineGrantID != sessionID {
		return "", oauth.ErrUnmatchedSession
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
		oauth.GrantSessionKindOffline,
		newRefreshTokenResult.TokenHash,
	)

	if err != nil {
		return "", err
	}

	return result.Token, nil
}
