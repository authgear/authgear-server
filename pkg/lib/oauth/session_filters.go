package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type SessionFilter interface {
	Keep(sess session.Session) bool
}

type SessionFilterFunc func(sess session.Session) bool

func (f SessionFilterFunc) Keep(sess session.Session) bool {
	return f(sess)
}

func ApplySessionFilters(sessions []session.Session, filters ...SessionFilter) (out []session.Session) {
	for _, sess := range sessions {
		keep := true
		for _, f := range filters {
			if !f.Keep(sess) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, sess)
		}
	}
	return
}

type RemoveThirdPartySessionFilter struct {
	ThirdPartyClientIDSet setutil.Set[string]
}

func NewRemoveThirdPartySessionFilter(oauthConfig *config.OAuthConfig) *RemoveThirdPartySessionFilter {
	s := make(setutil.Set[string])
	for _, c := range oauthConfig.Clients {
		if c.ClientParty() == config.ClientPartyThird {
			s[c.ClientID] = struct{}{}
		}
	}

	return &RemoveThirdPartySessionFilter{
		ThirdPartyClientIDSet: s,
	}
}

func (f *RemoveThirdPartySessionFilter) Keep(sess session.Session) bool {
	if offlineGrant, ok := sess.(*OfflineGrant); ok {
		if _, isThirdParty := f.ThirdPartyClientIDSet[offlineGrant.ClientID]; isThirdParty {
			return false
		}
	}
	return true
}

var _ SessionFilter = &RemoveThirdPartySessionFilter{}
