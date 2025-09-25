package handler

import (
	"context"

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
	ctx context.Context,
	options *IssuePreAuthenticatedURLTokenOptions,
) (*IssuePreAuthenticatedURLTokenResult, error) {
	now := s.Clock.NowUTC()
	token := oauth.GenerateToken()
	tokenHash := oauth.HashToken(token)
	err := s.PreAuthenticatedURLTokens.CreatePreAuthenticatedURLToken(ctx, &oauth.PreAuthenticatedURLToken{
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
	ctx context.Context,
	client *config.OAuthClientConfig,
	sessionID string,
	token string,
) (string, error) {
	tokenHash := oauth.HashToken(token)
	tokenModel, err := s.PreAuthenticatedURLTokens.ConsumePreAuthenticatedURLToken(ctx, tokenHash)
	if err != nil {
		return "", err
	}
	if tokenModel.ClientID != client.ClientID {
		return "", oauth.ErrUnmatchedClient
	}
	if tokenModel.OfflineGrantID != sessionID {
		return "", oauth.ErrUnmatchedSession
	}

	offlineGrant, err := s.OfflineGrantService.GetOfflineGrant(ctx, tokenModel.OfflineGrantID)
	if err != nil {
		return "", err
	}

	// DPoP is not important here, because the refresh token is not exposed
	dpopJKT := ""

	now := s.Clock.NowUTC()
	shortLivedRefreshTokenExpireAt := now.Add(client.AccessTokenLifetime.Duration())

	newRefreshTokenResult, newOfflineGrant, err := s.OfflineGrantService.CreateNewRefreshToken(ctx, oauth.CreateNewRefreshTokenOptions{
		OfflineGrant:                   offlineGrant,
		ClientID:                       tokenModel.ClientID,
		Scopes:                         tokenModel.Scopes,
		AuthorizationID:                tokenModel.AuthorizationID,
		DPoPJKT:                        dpopJKT,
		ShortLivedRefreshTokenExpireAt: &shortLivedRefreshTokenExpireAt,
	})
	if err != nil {
		return "", err
	}
	offlineGrant = newOfflineGrant

	issueAccessGrantOptions := oauth.IssueAccessGrantOptions{
		ClientConfig:            client,
		Scopes:                  tokenModel.Scopes,
		AuthorizationID:         tokenModel.AuthorizationID,
		AuthenticationInfo:      offlineGrant.GetAuthenticationInfo(),
		SessionLike:             offlineGrant,
		InitialRefreshTokenHash: newRefreshTokenResult.TokenHash,
	}
	result, err := s.AccessGrantService.IssueAccessGrant(
		ctx,
		issueAccessGrantOptions,
	)

	if err != nil {
		return "", err
	}

	return result.Token, nil
}
