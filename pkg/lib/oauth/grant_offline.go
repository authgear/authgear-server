package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

type OfflineGrant struct {
	AppID           string                 `json:"app_id"`
	ID              string                 `json:"id"`
	Labels          map[string]interface{} `json:"labels"`
	ClientID        string                 `json:"client_id"`
	AuthorizationID string                 `json:"authz_id"`
	IDPSessionID    string                 `json:"idp_session_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	Scopes    []string  `json:"scopes"`
	TokenHash string    `json:"token_hash"`

	Attrs      session.Attrs `json:"attrs"`
	AccessInfo access.Info   `json:"access_info"`

	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`
}

var _ Grant = &OfflineGrant{}

func (g *OfflineGrant) Session() (kind GrantSessionKind, id string) {
	return GrantSessionKindOffline, g.ID
}

func (g *OfflineGrant) SessionID() string            { return g.ID }
func (g *OfflineGrant) SessionType() session.Type    { return session.TypeOfflineGrant }
func (g *OfflineGrant) SessionAttrs() *session.Attrs { return &g.Attrs }

func (g *OfflineGrant) GetCreatedAt() time.Time     { return g.CreatedAt }
func (g *OfflineGrant) GetClientID() string         { return g.ClientID }
func (g *OfflineGrant) GetAccessInfo() *access.Info { return &g.AccessInfo }

func (g *OfflineGrant) ToAPIModel() *model.Session {
	var displayName string
	formattedDeviceInfo := deviceinfo.Format(g.DeviceInfo)
	ua := model.ParseUserAgent(g.AccessInfo.LastAccess.UserAgent)
	if formattedDeviceInfo != "" {
		displayName = formattedDeviceInfo
	} else {
		displayName = ua.Format()
	}

	amr, _ := g.Attrs.GetAMR()
	acr, _ := g.Attrs.GetACR()
	return &model.Session{
		Meta: model.Meta{
			ID:        g.ID,
			CreatedAt: g.CreatedAt,
			// TODO(session): Session Updated At should be the time user actively updates it.
			UpdatedAt: g.AccessInfo.LastAccess.Timestamp,
		},
		Type: model.SessionTypeOfflineGrant,

		AMR: amr,
		ACR: acr,

		LastAccessedAt:   g.AccessInfo.LastAccess.Timestamp,
		CreatedByIP:      g.AccessInfo.InitialAccess.RemoteIP,
		LastAccessedByIP: g.AccessInfo.LastAccess.RemoteIP,

		DisplayName: displayName,
	}
}
