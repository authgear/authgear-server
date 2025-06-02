package handler

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
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
	DPoPJKT              string
	IdentitySpecs        []*identity.Spec
}

func (s *CodeGrantService) CreateCodeGrant(ctx context.Context, opts *CreateCodeGrantOptions) (code string, grant *oauth.CodeGrant, err error) {
	code = s.CodeGenerator()
	codeHash := oauth.HashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:              string(s.AppID),
		AuthorizationID:    opts.Authorization.ID,
		AuthenticationInfo: opts.AuthenticationInfo,
		IDTokenHintSID:     opts.IDTokenHintSID,

		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(CodeGrantValidDuration),
		CodeHash:  codeHash,
		DPoPJKT:   opts.DPoPJKT,

		RedirectURI:          opts.RedirectURI,
		AuthorizationRequest: opts.AuthorizationRequest,
		IdentitySpecs:        opts.IdentitySpecs,
	}

	err = s.CodeGrants.CreateCodeGrant(ctx, codeGrant)
	if err != nil {
		return "", nil, err
	}
	otelauthgear.IntCounterAddOne(
		ctx,
		otelauthgear.CounterOAuthAuthorizationCodeCreationCount,
	)

	return code, codeGrant, nil
}
