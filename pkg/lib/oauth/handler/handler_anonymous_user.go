package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

var ErrUnauthenticated = apierrors.NewUnauthorized("authentication required")
var ErrLoggedInAsNormalUser = apierrors.NewInvalid("user logged in as normal user")

const PromotionCodeDuration = duration.Short

type anonymousSignupWithoutKeyInput struct{}

func (i *anonymousSignupWithoutKeyInput) GetAnonymousRequestToken() string { return "" }

func (i *anonymousSignupWithoutKeyInput) SignUpAnonymousUserWithoutKey() bool { return true }

func (i *anonymousSignupWithoutKeyInput) GetPromotionCode() string {
	return ""
}

var _ nodes.InputUseIdentityAnonymous = &anonymousSignupWithoutKeyInput{}

type UserProvider interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type AnonymousIdentityProvider interface {
	List(ctx context.Context, userID string) ([]*identity.Anonymous, error)
}

type PromotionCodeStore interface {
	CreatePromotionCode(ctx context.Context, code *anonymous.PromotionCode) error
}

type CookiesGetter interface {
	GetCookies() []*http.Cookie
}

type SignupAnonymousUserResult struct {
	TokenResponse interface{}
	Cookies       []*http.Cookie
}

type AnonymousUserHandler struct {
	AppID       config.AppID
	OAuthConfig *config.OAuthConfig

	Graphs              GraphService
	Authorizations      AuthorizationService
	Clock               clock.Clock
	TokenService        TokenService
	UserProvider        UserProvider
	AnonymousIdentities AnonymousIdentityProvider
	PromotionCodes      PromotionCodeStore
	OAuthClientResolver OAuthClientResolver
}

// SignupAnonymousUser return token response or api errors
func (h *AnonymousUserHandler) SignupAnonymousUser(
	ctx context.Context,
	req *http.Request,
	clientID string,
	sessionType WebSessionType,
	refreshToken string,
) (*SignupAnonymousUserResult, error) {
	switch sessionType {
	case WebSessionTypeCookie:
		return h.signupAnonymousUserWithCookieSessionType(ctx, req)
	case WebSessionTypeRefreshToken:
		return h.signupAnonymousUserWithRefreshTokenSessionType(ctx, req, clientID, refreshToken)
	default:
		panic("unknown web session type")
	}
}

func (h *AnonymousUserHandler) signupAnonymousUserWithCookieSessionType(
	ctx context.Context,
	req *http.Request,
) (*SignupAnonymousUserResult, error) {
	s := session.GetSession(ctx)
	if s != nil && s.SessionType() == session.TypeIdentityProvider {
		user, err := h.UserProvider.Get(ctx, s.GetAuthenticationInfo().UserID, accesscontrol.RoleGreatest)
		if err != nil {
			return nil, err
		}

		if user.IsAnonymous {
			return &SignupAnonymousUserResult{}, nil
		}
		return nil, ErrLoggedInAsNormalUser
	}

	graph, err := h.runSignupAnonymousUserGraph(ctx, false)
	if err != nil {
		return nil, err
	}

	cookies := []*http.Cookie{}
	for _, node := range graph.Nodes {
		if a, ok := node.(CookiesGetter); ok {
			cookies = append(cookies, a.GetCookies()...)
		}
	}

	return &SignupAnonymousUserResult{
		Cookies: cookies,
	}, nil
}

func (h *AnonymousUserHandler) signupAnonymousUserWithRefreshTokenSessionType(
	ctx context.Context,
	req *http.Request,
	clientID string,
	refreshToken string,
) (*SignupAnonymousUserResult, error) {
	ctx, client := resolveClient(ctx, h.OAuthClientResolver, clientID)
	if client == nil {
		// "invalid_client"
		return nil, apierrors.NewInvalid("invalid client ID")
	}

	if !client.HasFullAccessScope() {
		// unauthorized_client
		return nil, apierrors.NewInvalid("Anonymous user is not supported by the client application type. Try using SPA, Traditional Web App, or Native App client types if applicable.")
	}

	// TODO(oauth): allow specifying scopes for anonymous user signup
	scopes := []string{"openid", oauth.OfflineAccess, oauth.FullAccessScope}

	if refreshToken != "" {
		authz, grant, refreshTokenHash, err := h.TokenService.ParseRefreshToken(ctx, refreshToken)
		if errors.Is(err, ErrInvalidRefreshToken) {
			return nil, apierrors.NewInvalid("invalid refresh token")
		} else if err != nil {
			return nil, err
		}

		user, err := h.UserProvider.Get(ctx, authz.UserID, accesscontrol.RoleGreatest)
		if err != nil {
			return nil, err
		}
		if !user.IsAnonymous {
			return nil, ErrLoggedInAsNormalUser
		}

		resp := protocol.TokenResponse{}
		issueAccessGrantOptions := oauth.IssueAccessGrantOptions{
			ClientConfig:       client,
			Scopes:             scopes,
			AuthorizationID:    authz.ID,
			AuthenticationInfo: grant.GetAuthenticationInfo(),
			SessionLike:        grant,
			RefreshTokenHash:   refreshTokenHash,
		}
		err = h.TokenService.IssueAccessGrantByRefreshToken(ctx, issueAccessGrantOptions, resp)
		if err != nil {
			return nil, err
		}

		return &SignupAnonymousUserResult{
			TokenResponse: resp,
		}, nil
	}

	graph, err := h.runSignupAnonymousUserGraph(ctx, true)
	if err != nil {
		return nil, err
	}

	info := authenticationinfo.T{
		UserID:          graph.MustGetUserID(),
		AuthenticatedAt: h.Clock.NowUTC(),
	}

	authz, err := h.Authorizations.CheckAndGrant(
		ctx,
		client.ClientID,
		info.UserID,
		scopes,
	)
	if err != nil {
		return nil, err
	}

	dpopJKT, _ := dpop.GetDPoPProofJKT(ctx)

	resp := protocol.TokenResponse{}
	// SSOEnabled is false for refresh tokens that are granted by anonymous login
	opts := IssueOfflineGrantOptions{
		Scopes:             scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: info,
		DeviceInfo:         nil,
		SSOEnabled:         false,
		DPoPJKT:            dpopJKT,
	}
	offlineGrant, tokenHash, err := h.TokenService.IssueOfflineGrant(ctx, client, opts, resp)
	if err != nil {
		return nil, err
	}

	issueAccessGrantOptions := oauth.IssueAccessGrantOptions{
		ClientConfig:       client,
		Scopes:             scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: info,
		SessionLike:        offlineGrant,
		RefreshTokenHash:   tokenHash,
	}
	err = h.TokenService.IssueAccessGrantByRefreshToken(ctx, issueAccessGrantOptions, resp)
	if err != nil {
		return nil, err
	}

	return &SignupAnonymousUserResult{
		TokenResponse: resp,
	}, nil
}

func (h *AnonymousUserHandler) runSignupAnonymousUserGraph(
	ctx context.Context,
	suppressIDPSessionCookie bool,
) (*interaction.Graph, error) {
	var graph *interaction.Graph
	err := h.Graphs.DryRun(ctx, interaction.ContextValues{}, func(ctx context.Context, interactionCtx *interaction.Context) (*interaction.Graph, error) {
		var err error
		intent := &interactionintents.IntentAuthenticate{
			Kind:                     interactionintents.IntentAuthenticateKindLogin,
			SuppressIDPSessionCookie: suppressIDPSessionCookie,
		}
		graph, err = h.Graphs.NewGraph(ctx, interactionCtx, intent)
		if err != nil {
			return nil, err
		}

		var edges []interaction.Edge
		graph, edges, err = h.Graphs.Accept(ctx, interactionCtx, graph, &anonymousSignupWithoutKeyInput{})
		if len(edges) != 0 {
			return nil, errors.New("interaction not completed for anonymous users")
		} else if err != nil {
			return nil, err
		}

		return graph, nil
	})

	if apierrors.IsKind(err, api.InvariantViolated) &&
		apierrors.AsAPIError(err).HasCause("AnonymousUserDisallowed") {
		// anonymous user not enabled
		return nil, apierrors.NewInvalid("anonymous user disallowed")
	} else if errors.Is(err, api.ErrInvalidCredentials) {
		// invalid_grant
		return nil, apierrors.NewInvalid(api.InvalidCredentials.Reason)
	} else if err != nil {
		return nil, err
	}

	err = h.Graphs.Run(ctx, interaction.ContextValues{}, graph)
	if apierrors.IsAPIError(err) {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return graph, nil
}

func (h *AnonymousUserHandler) IssuePromotionCode(
	ctx context.Context,
	req *http.Request,
	sessionType WebSessionType,
	refreshToken string,
) (code string, codeObj *anonymous.PromotionCode, err error) {
	var appID, userID string
	switch sessionType {
	case WebSessionTypeRefreshToken:
		if refreshToken == "" {
			err = ErrUnauthenticated
			return
		}
		authz, _, _, e := h.TokenService.ParseRefreshToken(ctx, refreshToken)
		var oauthError *protocol.OAuthProtocolError
		if errors.As(e, &oauthError) {
			err = apierrors.NewForbidden(oauthError.Error())
			return
		} else if e != nil {
			err = e
			return
		}
		// Ensure client is authorized with full user access (i.e. first-party client)
		if !authz.IsAuthorized([]string{oauth.FullAccessScope}) {
			err = apierrors.NewForbidden("the client is not authorized to have full user access")
			return
		}

		appID = authz.AppID
		userID = authz.UserID
	case WebSessionTypeCookie:
		s := session.GetSession(ctx)
		if s != nil && s.SessionType() == session.TypeIdentityProvider {
			appID = string(h.AppID)
			userID = s.GetAuthenticationInfo().UserID
		} else {
			err = ErrUnauthenticated
			return
		}
	default:
		panic("unknown web session type")
	}

	user, err := h.UserProvider.Get(ctx, userID, accesscontrol.RoleGreatest)
	if err != nil {
		return
	}
	if !user.IsAnonymous {
		err = ErrLoggedInAsNormalUser
		return
	}

	identities, err := h.AnonymousIdentities.List(ctx, userID)
	if err != nil {
		return
	}
	if len(identities) != 1 {
		panic(fmt.Errorf("api: expected has 1 anonymous identity for anonymous user, got %d", len(identities)))
	}

	now := h.Clock.NowUTC()
	c := anonymous.GeneratePromotionCode()
	cObj := &anonymous.PromotionCode{
		AppID:      appID,
		UserID:     userID,
		IdentityID: identities[0].ID,
		CreatedAt:  now,
		ExpireAt:   now.Add(PromotionCodeDuration),
		CodeHash:   anonymous.HashPromotionCode(c),
	}
	err = h.PromotionCodes.CreatePromotionCode(ctx, cObj)
	if err != nil {
		return
	}
	code = c
	codeObj = cObj
	return
}
