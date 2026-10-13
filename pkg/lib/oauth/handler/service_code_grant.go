package handler

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
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
	RedirectURI          string
	AuthorizationRequest protocol.AuthorizationRequest
	DPoPJKT              string
}

func (s *CodeGrantService) CreateCodeGrant(ctx context.Context, opts *CreateCodeGrantOptions) (code string, grant *oauth.CodeGrant, err error) {
	code = s.CodeGenerator()
	codeHash := oauth.HashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:              string(s.AppID),
		AuthorizationID:    opts.Authorization.ID,
		AuthenticationInfo: opts.AuthenticationInfo,
		// Populated from the trusted sid on the authentication info -- never from
		// the OAuth session directly. Empty unless a verified id_token_hint named
		// this user's session (set in the authorize/consent handlers).
		IDTokenHintSID: opts.AuthenticationInfo.IDTokenHintSID,

		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(CodeGrantValidDuration),
		CodeHash:  codeHash,
		DPoPJKT:   opts.DPoPJKT,

		RedirectURI:          opts.RedirectURI,
		AuthorizationRequest: opts.AuthorizationRequest,
		IdentitySpecs:        opts.AuthenticationInfo.IdentitySpecs,
	}

	err = s.CodeGrants.CreateCodeGrant(ctx, codeGrant)
	if err != nil {
		return "", nil, err
	}
	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterOAuthAuthorizationCodeCreationCount,
	)

	return code, codeGrant, nil
}
