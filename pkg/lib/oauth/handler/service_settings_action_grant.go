package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type SettingsActionGrantService struct {
	AppID         config.AppID
	CodeGenerator TokenGenerator
	Clock         clock.Clock

	SettingsActionGrants oauth.SettingsActionGrantStore
}

type CreateSettingsActionGrantOptions struct {
	Authorization        *oauth.Authorization
	IDPSessionID         string
	AuthenticationInfo   authenticationinfo.T
	IDTokenHintSID       string
	RedirectURI          string
	AuthorizationRequest protocol.AuthorizationRequest
}

func (s *SettingsActionGrantService) CreateSettingsActionGrant(opts *CreateSettingsActionGrantOptions) (code string, grant *oauth.SettingsActionGrant, err error) {
	code = s.CodeGenerator()
	codeHash := oauth.HashToken(code)

	settingsActionGrant := &oauth.SettingsActionGrant{
		AppID:              string(s.AppID),
		AuthorizationID:    opts.Authorization.ID,
		IDPSessionID:       opts.IDPSessionID,
		AuthenticationInfo: opts.AuthenticationInfo,
		IDTokenHintSID:     opts.IDTokenHintSID,

		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(SettingsActionGrantValidDuration),
		CodeHash:  codeHash,

		RedirectURI:          opts.RedirectURI,
		AuthorizationRequest: opts.AuthorizationRequest,
	}

	err = s.SettingsActionGrants.CreateSettingsActionGrant(settingsActionGrant)
	if err != nil {
		return "", nil, err
	}
	return code, settingsActionGrant, nil
}
