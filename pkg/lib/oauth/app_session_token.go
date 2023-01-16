package oauth

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type AppSessionToken struct {
	AppID          string `json:"app_id"`
	OfflineGrantID string `json:"offline_grant_id"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	TokenHash string    `json:"token_hash"`
}

type AppSessionTokenResult struct {
	Cookies     []*http.Cookie
	RedirectURI string
}

func (r *AppSessionTokenResult) WriteResponse(w http.ResponseWriter, req *http.Request) {
	for _, cookie := range r.Cookies {
		httputil.UpdateCookie(w, cookie)
	}
	http.Redirect(w, req, r.RedirectURI, http.StatusFound)
}

func (r *AppSessionTokenResult) IsInternalError() bool {
	return false
}

type AppSessionTokenServiceOfflineGrantService interface {
	IsValid(session *OfflineGrant) (valid bool, expiry time.Time, err error)
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
	OfflineGrants       OfflineGrantStore
	OfflineGrantService AppSessionTokenServiceOfflineGrantService
	Cookies             AppSessionTokenServiceCookieManager
	Clock               clock.Clock
}

func (s *AppSessionTokenService) Handle(input AppSessionTokenInput) (httputil.Result, error) {
	token, err := s.exchange(input.AppSessionToken)
	if err != nil {
		return nil, err
	}

	cookie := s.Cookies.ValueCookie(session.AppSessionTokenCookieDef, token)
	return &AppSessionTokenResult{
		Cookies:     []*http.Cookie{cookie},
		RedirectURI: input.RedirectURI,
	}, nil
}

func (s *AppSessionTokenService) exchange(appSessionToken string) (string, error) {
	sToken, err := s.AppSessionTokens.GetAppSessionToken(HashToken(appSessionToken))
	if err != nil {
		return "", err
	}

	offlineGrant, err := s.OfflineGrants.GetOfflineGrant(sToken.OfflineGrantID)
	if err != nil {
		return "", err
	}

	isValid, expiry, err := s.OfflineGrantService.IsValid(offlineGrant)
	if err != nil {
		return "", err
	}

	if !isValid {
		return "", ErrGrantNotFound
	}

	err = s.AppSessionTokens.DeleteAppSessionToken(sToken)
	if err != nil {
		return "", err
	}

	// Create app session
	token := GenerateToken()
	appSession := &AppSession{
		AppID:          offlineGrant.AppID,
		OfflineGrantID: offlineGrant.ID,
		CreatedAt:      s.Clock.NowUTC(),
		ExpireAt:       expiry,
		TokenHash:      HashToken(token),
	}
	err = s.AppSessions.CreateAppSession(appSession)
	if err != nil {
		return "", err
	}

	return token, nil
}
