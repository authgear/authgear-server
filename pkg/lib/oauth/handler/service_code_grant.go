package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type CodeGrantService struct {
	AppID         config.AppID
	CodeGenerator TokenGenerator
	Clock         clock.Clock

	CodeGrants oauth.CodeGrantStore
}

type CreateCodeGrantOptions struct {
	Authorization        *oauth.Authorization
	SessionType          session.Type
	SessionID            string
	AuthenticationInfo   authenticationinfo.T
	IDTokenHintSID       string
	RedirectURI          string
	AuthorizationRequest protocol.AuthorizationRequest
}

func (s *CodeGrantService) CreateCodeGrant(opts *CreateCodeGrantOptions) (code string, grant *oauth.CodeGrant, err error) {
	code = s.CodeGenerator()
	codeHash := oauth.HashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:              string(s.AppID),
		AuthorizationID:    opts.Authorization.ID,
		SessionType:        opts.SessionType,
		SessionID:          opts.SessionID,
		AuthenticationInfo: opts.AuthenticationInfo,
		IDTokenHintSID:     opts.IDTokenHintSID,

		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(CodeGrantValidDuration),
		CodeHash:  codeHash,

		RedirectURI:          opts.RedirectURI,
		AuthorizationRequest: opts.AuthorizationRequest,
	}

	err = s.CodeGrants.CreateCodeGrant(codeGrant)
	if err != nil {
		return "", nil, err
	}
	return code, codeGrant, nil
}
