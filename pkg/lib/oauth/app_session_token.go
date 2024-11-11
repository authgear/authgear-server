package oauth

import (
	"context"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type AppSessionToken struct {
	AppID          string `json:"app_id"`
	OfflineGrantID string `json:"offline_grant_id"`

	CreatedAt        time.Time `json:"created_at"`
	ExpireAt         time.Time `json:"expire_at"`
	TokenHash        string    `json:"token_hash"`
	RefreshTokenHash string    `json:"refresh_token_hash"`
}

type AppSessionTokenServiceOfflineGrantService interface {
	GetOfflineGrant(ctx context.Context, id string) (*OfflineGrant, error)
}

type AppSessionTokenServiceCookieManager interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type AppSessionTokenInput struct {
	AppSessionToken string
	RedirectURI     string
}

type AppSessionTokenService struct {
	AppSessions         AppSessionStore
	AppSessionTokens    AppSessionTokenStore
	OfflineGrantService AppSessionTokenServiceOfflineGrantService
	Cookies             AppSessionTokenServiceCookieManager
	Clock               clock.Clock
}

func (s *AppSessionTokenService) Handle(ctx context.Context, input AppSessionTokenInput) (httputil.Result, error) {
	token, err := s.Exchange(ctx, input.AppSessionToken)
	if err != nil {
		return nil, err
	}

	cookie := s.Cookies.ValueCookie(session.AppSessionTokenCookieDef, token)
	return &httputil.ResultRedirect{
		Cookies: []*http.Cookie{cookie},
		URL:     input.RedirectURI,
	}, nil
}

func (s *AppSessionTokenService) Exchange(ctx context.Context, appSessionToken string) (string, error) {
	sToken, err := s.AppSessionTokens.GetAppSessionToken(ctx, HashToken(appSessionToken))
	if err != nil {
		return "", err
	}
	refreshTokenHash := sToken.RefreshTokenHash

	offlineGrant, err := s.OfflineGrantService.GetOfflineGrant(ctx, sToken.OfflineGrantID)
	if err != nil {
		return "", err
	}

	err = s.AppSessionTokens.DeleteAppSessionToken(ctx, sToken)
	if err != nil {
		return "", err
	}

	// Create app session
	token := GenerateToken()
	appSession := &AppSession{
		AppID:            offlineGrant.AppID,
		OfflineGrantID:   offlineGrant.ID,
		CreatedAt:        s.Clock.NowUTC(),
		ExpireAt:         offlineGrant.ExpireAtForResolvedSession,
		TokenHash:        HashToken(token),
		RefreshTokenHash: refreshTokenHash,
	}
	err = s.AppSessions.CreateAppSession(ctx, appSession)
	if err != nil {
		return "", err
	}

	return token, nil
}
