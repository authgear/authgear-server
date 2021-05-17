package sessiongroup

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

type keyDeviceModelDeviceName struct {
	DeviceModel string
	DeviceName  string
}

type group struct {
	OfflineGrants []*oauth.OfflineGrant
	IDPSessions   []*idpsession.IDPSession
}

// Group groups sessions into a list of SessionGroup.
func Group(sessions []session.Session) []*model.SessionGroup {
	// 1. Offline grants are first grouped by Device Name and Device model.
	groupMap := make(map[keyDeviceModelDeviceName]group)
	for _, sess := range sessions {
		if offlineGrant, ok := sess.(*oauth.OfflineGrant); ok {
			if deviceInfo, ok := offlineGrant.GetDeviceInfo(); ok {
				key := keyDeviceModelDeviceName{
					DeviceModel: deviceinfo.DeviceModel(deviceInfo),
					DeviceName:  deviceinfo.DeviceName(deviceInfo),
				}
				group := groupMap[key]
				group.OfflineGrants = append(group.OfflineGrants, offlineGrant)
				groupMap[key] = group
			}
		}
	}

	// 2. IdP sessions referenced by an offline grant are grouped together.
	//    IdP sessions referenced by an offline grant may appear in more than 1 group.
	var ungroupedIDPSessions []*idpsession.IDPSession
	for _, sess := range sessions {
		if idpSession, ok := sess.(*idpsession.IDPSession); ok {
			grouped := false
			for key, group := range groupMap {
				for _, offlineGrant := range group.OfflineGrants {
					if offlineGrant.IDPSessionID == idpSession.ID {
						grouped = true
						group.IDPSessions = append(group.IDPSessions, idpSession)
						groupMap[key] = group
					}
				}
			}
			if !grouped {
				ungroupedIDPSessions = append(ungroupedIDPSessions, idpSession)
			}
		}
	}

	var out []*model.SessionGroup
	for key, group := range groupMap {
		sessionGroup := &model.SessionGroup{
			Type:        model.SessionGroupTypeGrouped,
			DisplayName: key.DeviceModel,
		}
		for _, offlineGrant := range group.OfflineGrants {
			sessionGroup.OfflineGrantIDs = append(sessionGroup.OfflineGrantIDs, offlineGrant.ID)
			sessionGroup.Sessions = append(sessionGroup.Sessions, offlineGrant.ToAPIModel())
		}
		for _, idpSession := range group.IDPSessions {
			sessionGroup.Sessions = append(sessionGroup.Sessions, idpSession.ToAPIModel())
		}
		sortSessions(sessionGroup.Sessions)
		sessionGroup.LastAccessedAt = sessionGroup.Sessions[0].LastAccessedAt

		out = append(out, sessionGroup)
	}

	for _, idpSession := range ungroupedIDPSessions {
		apiModel := idpSession.ToAPIModel()
		sessions := []*model.Session{apiModel}
		lastAccessedAt := apiModel.LastAccessedAt
		out = append(out, &model.SessionGroup{
			Type:           model.SessionGroupTypeUngrouped,
			DisplayName:    apiModel.DisplayName,
			LastAccessedAt: lastAccessedAt,
			Sessions:       sessions,
		})
	}

	sortGroups(out)
	return out
}

func sortSessions(sessions []*model.Session) {
	sort.Slice(sessions, func(i, j int) bool {
		a := sessions[i]
		b := sessions[j]
		return a.LastAccessedAt.After(b.LastAccessedAt)
	})
}

func sortGroups(groups []*model.SessionGroup) {
	sort.Slice(groups, func(i, j int) bool {
		a := groups[i]
		b := groups[j]
		return a.LastAccessedAt.After(b.LastAccessedAt)
	})
}
