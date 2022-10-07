package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type AuthorizationFilter interface {
	Keep(authz *Authorization) bool
}

type AuthorizationFilterFunc func(a *Authorization) bool

func (f AuthorizationFilterFunc) Keep(a *Authorization) bool {
	return f(a)
}

func ApplyAuthorizationFilters(authzs []*Authorization, filters ...AuthorizationFilter) (out []*Authorization) {
	for _, authz := range authzs {
		keep := true
		for _, f := range filters {
			if !f.Keep(authz) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, authz)
		}
	}
	return
}

type KeepThirdPartyAuthorizationFilter struct {
	ThirdPartyClientIDSet setutil.Set[string]
}

func NewKeepThirdPartyAuthorizationFilter(oauthConfig *config.OAuthConfig) *KeepThirdPartyAuthorizationFilter {
	s := make(setutil.Set[string])
	for _, c := range oauthConfig.Clients {
		if c.ClientParty() == config.ClientPartyThird {
			s[c.ClientID] = struct{}{}
		}
	}

	return &KeepThirdPartyAuthorizationFilter{
		ThirdPartyClientIDSet: s,
	}
}

func (f *KeepThirdPartyAuthorizationFilter) Keep(authz *Authorization) bool {
	if _, isThirdParty := f.ThirdPartyClientIDSet[authz.ClientID]; isThirdParty {
		return true
	}
	return false
}

var _ AuthorizationFilter = &KeepThirdPartyAuthorizationFilter{}
