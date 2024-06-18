package sessionlisting

import (
	"sort"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

//go:generate mockgen -source=listing.go -destination=listing_mock_test.go -package sessionlisting_test

// Session in the sessionlisting package wrapped the model.Session to provide
// extra information for internal display
type Session struct {
	*model.Session
	// IsDevice has a different meaning for IDP session and offline grant
	// For IDP session, IsDevice is true only if it has active sso enabled offline grant.
	// So it may change.
	// All offline grant's IsDevice is true.
	IsDevice bool `json:"-"`
	// IsCurrent indicates if the session is current session
	IsCurrent bool `json:"-"`
}

type IDPSessionProvider interface {
	CheckSessionExpired(session *idpsession.IDPSession) (expired bool)
}

type OfflineGrantService interface {
	CheckSessionExpired(session *oauth.OfflineGrant) (bool, time.Time, error)
}

type SessionListingService struct {
	OAuthConfig   *config.OAuthConfig
	IDPSessions   IDPSessionProvider
	OfflineGrants OfflineGrantService
}

func (s *SessionListingService) FilterForDisplay(sessions []session.ListableSession, currentSession session.Session) ([]*Session, error) {
	sess := make([]session.ListableSession, len(sessions))
	copy(sess, sessions)
	sortSessions(sess)

	offlineGrants := []*oauth.OfflineGrant{}
	idpSessions := []*idpsession.IDPSession{}
	for _, ses := range sess {
		if offlineGrant, ok := ses.(*oauth.OfflineGrant); ok {
			offlineGrants = append(offlineGrants, offlineGrant)
		} else if idpSession, ok := ses.(*idpsession.IDPSession); ok {
			idpSessions = append(idpSessions, idpSession)
		}
	}

	// construct third-party app client id set
	thirdPartyClientIDSet := make(setutil.Set[string])
	for _, c := range s.OAuthConfig.Clients {
		if c.IsThirdParty() {
			thirdPartyClientIDSet[c.ClientID] = struct{}{}
		}
	}

	result := []*Session{}
	idpSessionToDisplayNameMap := map[string]string{}

	for _, offlineGrant := range offlineGrants {
		// remove third-party app refresh token
		// TODO(DEV-1403): Check all client id?
		if _, ok := thirdPartyClientIDSet[offlineGrant.ClientID]; ok {
			continue
		}

		expired, _, err := s.OfflineGrants.CheckSessionExpired(offlineGrant)
		if err != nil {
			return nil, err
		}
		if expired {
			continue
		}

		apiModel := &Session{
			Session:  offlineGrant.ToAPIModel(),
			IsDevice: true,
		}
		// construct a map for replacing idp session's display name in the SSO group
		if offlineGrant.SSOGroupIDPSessionID() != "" {
			if _, ok := idpSessionToDisplayNameMap[offlineGrant.SSOGroupIDPSessionID()]; !ok {
				idpSessionToDisplayNameMap[offlineGrant.SSOGroupIDPSessionID()] = apiModel.DisplayName
			}
			continue
		}

		if currentSession != nil {
			apiModel.IsCurrent = offlineGrant.IsSameSSOGroup(currentSession)
		}

		result = append(result, apiModel)
	}

	for _, idpSession := range idpSessions {
		expired := s.IDPSessions.CheckSessionExpired(idpSession)
		if expired {
			continue
		}

		// replace idp session display name with the last accessed refresh token's display name
		apiModel := &Session{
			Session: idpSession.ToAPIModel(),
		}
		if displayName, ok := idpSessionToDisplayNameMap[idpSession.ID]; ok {
			apiModel.DisplayName = displayName
			apiModel.IsDevice = true
		}

		if currentSession != nil {
			apiModel.IsCurrent = idpSession.IsSameSSOGroup(currentSession)
		}

		result = append(result, apiModel)
	}

	sortSessionModels(result)
	return result, nil
}

func sortSessions(sessions []session.ListableSession) {
	sort.Slice(sessions, func(i, j int) bool {
		a := time.Time{}
		if accessInfo := sessions[i].GetAccessInfo(); accessInfo != nil {
			a = accessInfo.LastAccess.Timestamp
		}
		b := time.Time{}
		if accessInfo := sessions[j].GetAccessInfo(); accessInfo != nil {
			b = accessInfo.LastAccess.Timestamp
		}
		return a.After(b)
	})
}

func sortSessionModels(sessions []*Session) {
	sort.Slice(sessions, func(i, j int) bool {
		a := sessions[i]
		b := sessions[j]
		return a.LastAccessedAt.After(b.LastAccessedAt)
	})
}
