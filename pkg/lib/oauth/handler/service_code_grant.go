package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type CodeGrantService struct {
	AppID         config.AppID
	CodeGenerator TokenGenerator
	Clock         clock.Clock

	CodeGrants oauth.CodeGrantStore
}

type CreateCodeGrantOptions struct {
	Authorization      *oauth.Authorization
	IDPSessionID       string
	AuthenticationInfo authenticationinfo.T
	IDTokenHintSID     string
	Scopes             []string
	RedirectURI        string
	OIDCNonce          string
	PKCEChallenge      string
	SSOEnabled         bool
}

func (s *CodeGrantService) CreateCodeGrant(opts *CreateCodeGrantOptions) (code string, grant *oauth.CodeGrant, err error) {
	code = s.CodeGenerator()
	codeHash := oauth.HashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:              string(s.AppID),
		AuthorizationID:    opts.Authorization.ID,
		IDPSessionID:       opts.IDPSessionID,
		AuthenticationInfo: opts.AuthenticationInfo,
		IDTokenHintSID:     opts.IDTokenHintSID,

		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(CodeGrantValidDuration),
		Scopes:    opts.Scopes,
		CodeHash:  codeHash,

		RedirectURI:   opts.RedirectURI,
		OIDCNonce:     opts.OIDCNonce,
		PKCEChallenge: opts.PKCEChallenge,
		SSOEnabled:    opts.SSOEnabled,
	}

	err = s.CodeGrants.CreateCodeGrant(codeGrant)
	if err != nil {
		return "", nil, err
	}
	return code, codeGrant, nil
}
