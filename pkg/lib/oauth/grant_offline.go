package oauth

import (
	"crypto/subtle"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
	"github.com/authgear/authgear-server/pkg/util/geoip"
)

type OfflineGrantRefreshToken struct {
	TokenHash       string    `json:"token_hash"`
	ClientID        string    `json:"client_id"`
	CreatedAt       time.Time `json:"created_at"`
	Scopes          []string  `json:"scopes"`
	AuthorizationID string    `json:"authz_id"`
}

type OfflineGrant struct {
	AppID           string `json:"app_id"`
	ID              string `json:"id"`
	ClientID        string `json:"client_id"`
	AuthorizationID string `json:"authz_id"`
	// IDPSessionID refers to the IDP session.
	IDPSessionID string `json:"idp_session_id,omitempty"`
	// IdentityID refers to the identity.
	// It is only set for biometric authentication.
	IdentityID string `json:"identity_id,omitempty"`

	CreatedAt       time.Time `json:"created_at"`
	AuthenticatedAt time.Time `json:"authenticated_at"`
	Scopes          []string  `json:"scopes"`
	TokenHash       string    `json:"token_hash"`

	Attrs      session.Attrs `json:"attrs"`
	AccessInfo access.Info   `json:"access_info"`

	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`

	SSOEnabled bool `json:"sso_enabled,omitempty"`

	App2AppDeviceKeyJWKJSON string `json:"app2app_device_key_jwk_json"`

	RefreshTokens []OfflineGrantRefreshToken `json:"refresh_tokens,omitempty"`
}

var _ session.ListableSession = &OfflineGrant{}

type OfflineGrantSession struct {
	OfflineGrant *OfflineGrant
	CreatedAt    time.Time
	TokenHash    string
	ClientID     string
	Scopes       []string
}

func (o *OfflineGrantSession) Session() {}
func (o *OfflineGrantSession) SessionID() string {
	return o.OfflineGrant.ID
}
func (o *OfflineGrantSession) SessionType() session.Type {
	return o.OfflineGrant.SessionType()
}
func (o *OfflineGrantSession) GetCreatedAt() time.Time {
	return o.CreatedAt
}
func (o *OfflineGrantSession) GetAuthenticationInfo() authenticationinfo.T {
	return o.OfflineGrant.GetAuthenticationInfo()
}
func (o *OfflineGrantSession) GetAccessInfo() *access.Info {
	return &o.OfflineGrant.AccessInfo
}
func (o *OfflineGrantSession) SSOGroupIDPSessionID() string {
	return o.OfflineGrant.SSOGroupIDPSessionID()
}

var _ session.Session = &OfflineGrantSession{}

func (g *OfflineGrant) ListableSession()          {}
func (g *OfflineGrant) SessionID() string         { return g.ID }
func (g *OfflineGrant) SessionType() session.Type { return session.TypeOfflineGrant }

func (g *OfflineGrant) GetCreatedAt() time.Time                       { return g.CreatedAt }
func (g *OfflineGrant) GetAuthenticatedAt() time.Time                 { return g.AuthenticatedAt }
func (g *OfflineGrant) GetClientID() string                           { return g.ClientID }
func (g *OfflineGrant) GetAccessInfo() *access.Info                   { return &g.AccessInfo }
func (g *OfflineGrant) GetDeviceInfo() (map[string]interface{}, bool) { return g.DeviceInfo, true }
func (g *OfflineGrant) GetUserID() string                             { return g.Attrs.UserID }
func (g *OfflineGrant) GetOIDCAMR() ([]string, bool)                  { return g.Attrs.GetAMR() }

func (g *OfflineGrant) ToAPIModel() *model.Session {
	var displayName string
	formattedDeviceInfo := deviceinfo.DeviceModel(g.DeviceInfo)
	ua := model.ParseUserAgent(g.AccessInfo.LastAccess.UserAgent)
	if formattedDeviceInfo != "" {
		displayName = formattedDeviceInfo
	} else {
		displayName = ua.Format()
	}

	amr, _ := g.Attrs.GetAMR()

	apiModel := &model.Session{
		Meta: model.Meta{
			ID:        g.ID,
			CreatedAt: g.CreatedAt,
			// TODO(session): Session Updated At should be the time user actively updates it.
			UpdatedAt: g.AccessInfo.LastAccess.Timestamp,
		},
		Type: model.SessionTypeOfflineGrant,

		AMR:      amr,
		ClientID: &g.ClientID,

		LastAccessedAt:   g.AccessInfo.LastAccess.Timestamp,
		CreatedByIP:      g.AccessInfo.InitialAccess.RemoteIP,
		LastAccessedByIP: g.AccessInfo.LastAccess.RemoteIP,

		DisplayName:     displayName,
		ApplicationName: deviceinfo.ApplicationName(g.DeviceInfo),
		UserAgent:       ua.Raw,
	}

	ipInfo, ok := geoip.DefaultDatabase.IPString(g.AccessInfo.LastAccess.RemoteIP)
	if ok {
		apiModel.LastAccessedByIPCountryCode = ipInfo.CountryCode
		apiModel.LastAccessedByIPEnglishCountryName = ipInfo.EnglishCountryName
	}

	return apiModel
}

func (g *OfflineGrant) GetAuthenticationInfo() authenticationinfo.T {
	amr, _ := g.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:          g.GetUserID(),
		AuthenticatedAt: g.GetAuthenticatedAt(),
		AMR:             amr,
	}
}

func (g *OfflineGrant) SSOGroupIDPSessionID() string {
	if g.SSOEnabled {
		return g.IDPSessionID
	}
	return ""
}

// IsSameSSOGroup returns true when the session argument
// - is the same offline grant
// - is idp session in the same sso group (current offline grant needs to be sso enabled)
// - is offline grant in the same sso group (current offline grant needs to be sso enabled)
func (g *OfflineGrant) IsSameSSOGroup(ss session.SessionBase) bool {
	if g.EqualSession(ss) {
		return true
	}

	if g.SSOEnabled {
		if g.SSOGroupIDPSessionID() == "" {
			return false
		}
		return g.SSOGroupIDPSessionID() == ss.SSOGroupIDPSessionID()
	}

	return false
}

func (g *OfflineGrant) EqualSession(ss session.SessionBase) bool {
	return g.SessionID() == ss.SessionID() && g.SessionType() == ss.SessionType()
}

func (g *OfflineGrant) ToSession(refreshTokenHash string) session.Session {
	// Note(tung): For backward compatibility,
	// if refreshTokenHash is empty, the "root" offline grant should be used
	if refreshTokenHash == "" || refreshTokenHash == g.TokenHash {
		return &OfflineGrantSession{
			OfflineGrant: g,
			CreatedAt:    g.CreatedAt,
			TokenHash:    g.TokenHash,
			ClientID:     g.ClientID,
			Scopes:       g.Scopes,
		}
	}

	for _, token := range g.RefreshTokens {
		if token.TokenHash == refreshTokenHash {
			return &OfflineGrantSession{
				OfflineGrant: g,
				CreatedAt:    token.CreatedAt,
				TokenHash:    token.TokenHash,
				ClientID:     token.ClientID,
				Scopes:       token.Scopes,
			}
		}
	}

	panic("offline grant: cannot find a token with the provided token hash")
}

func (g *OfflineGrant) MatchHash(refreshTokenHash string) bool {
	// TODO(tung): Implement this function so that it has constant execution time?
	if subtle.ConstantTimeCompare([]byte(refreshTokenHash), []byte(g.TokenHash)) == 1 {
		return true
	}

	for _, token := range g.RefreshTokens {
		if subtle.ConstantTimeCompare([]byte(refreshTokenHash), []byte(token.TokenHash)) == 1 {
			return true
		}
	}

	return false
}
