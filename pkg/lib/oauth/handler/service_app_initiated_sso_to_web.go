package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type PreAuthenticatedURLTokenServiceImpl struct {
	Clock clock.Clock

	PreAuthenticatedURLTokens oauth.PreAuthenticatedURLTokenStore
	AccessGrantService        oauth.PreAuthenticatedURLTokenAccessGrantService
	OfflineGrantService       oauth.PreAuthenticatedURLTokenOfflineGrantService
}

type IssuePreAuthenticatedURLTokenResult struct {
	Token     string
	TokenHash string
	TokenType string
	ExpiresIn int
}

type IssuePreAuthenticatedURLTokenOptions struct {
	AppID           string
	ClientID        string
	OfflineGrantID  string
	AuthorizationID string
	Scopes          []string
}

func (s *PreAuthenticatedURLTokenServiceImpl) IssuePreAuthenticatedURLToken(
	options *IssuePreAuthenticatedURLTokenOptions,
) (*IssuePreAuthenticatedURLTokenResult, error) {
	now := s.Clock.NowUTC()
	token := oauth.GenerateToken()
	tokenHash := oauth.HashToken(token)
	err := s.PreAuthenticatedURLTokens.CreatePreAuthenticatedURLToken(&oauth.PreAuthenticatedURLToken{
		AppID:           options.AppID,
		AuthorizationID: options.AuthorizationID,
		ClientID:        options.ClientID,
		OfflineGrantID:  options.OfflineGrantID,
		Scopes:          options.Scopes,

		CreatedAt: now,
		ExpireAt:  now.Add(oauth.PreAuthenticatedURLTokenLifetime),
		TokenHash: tokenHash,
	})
	if err != nil {
		return nil, err
	}

	return &IssuePreAuthenticatedURLTokenResult{
		Token:     token,
		TokenHash: tokenHash,
		TokenType: "Bearer",
		ExpiresIn: int(oauth.PreAuthenticatedURLTokenLifetime.Seconds()),
	}, nil
}

func (s *PreAuthenticatedURLTokenServiceImpl) ExchangeForAccessToken(
	client *config.OAuthClientConfig,
	sessionID string,
	token string,
) (string, error) {
	tokenHash := oauth.HashToken(token)
	tokenModel, err := s.PreAuthenticatedURLTokens.ConsumePreAuthenticatedURLToken(tokenHash)
	if err != nil {
		return "", err
	}
	if tokenModel.ClientID != client.ClientID {
		return "", oauth.ErrUnmatchedClient
	}
	if tokenModel.OfflineGrantID != sessionID {
		return "", oauth.ErrUnmatchedSession
	}

	offlineGrant, err := s.OfflineGrantService.GetOfflineGrant(tokenModel.OfflineGrantID)
	if err != nil {
		return "", err
	}

	// DPoP is not important here, because the refresh token is not exposed
	dpopJKT := ""

	newRefreshTokenResult, newOfflineGrant, err := s.OfflineGrantService.CreateNewRefreshToken(
		offlineGrant, tokenModel.ClientID, tokenModel.Scopes, tokenModel.AuthorizationID, dpopJKT,
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
