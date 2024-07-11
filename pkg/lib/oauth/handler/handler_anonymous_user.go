package handler

import (
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
	"github.com/authgear/authgear-server/pkg/util/log"
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

type AnonymousUserHandlerLogger struct{ *log.Logger }

func NewAnonymousUserHandlerLogger(lf *log.Factory) AnonymousUserHandlerLogger {
	return AnonymousUserHandlerLogger{lf.New("oauth-anonymous-user")}
}

type UserProvider interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
}

type AnonymousIdentityProvider interface {
	List(userID string) ([]*identity.Anonymous, error)
}

type PromotionCodeStore interface {
	CreatePromotionCode(code *anonymous.PromotionCode) error
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
	Logger      AnonymousUserHandlerLogger

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
	req *http.Request,
	clientID string,
	sessionType WebSessionType,
	refreshToken string,
) (*SignupAnonymousUserResult, error) {
	switch sessionType {
	case WebSessionTypeCookie:
		return h.signupAnonymousUserWithCookieSessionType(req)
	case WebSessionTypeRefreshToken:
		return h.signupAnonymousUserWithRefreshTokenSessionType(req, clientID, refreshToken)
	default:
		panic("unknown web session type")
	}
}

func (h *AnonymousUserHandler) signupAnonymousUserWithCookieSessionType(
	req *http.Request,
) (*SignupAnonymousUserResult, error) {
	s := session.GetSession(req.Context())
	if s != nil && s.SessionType() == session.TypeIdentityProvider {
		user, err := h.UserProvider.Get(s.GetAuthenticationInfo().UserID, accesscontrol.RoleGreatest)
		if err != nil {
			return nil, err
		}

		if user.IsAnonymous {
			return &SignupAnonymousUserResult{}, nil
		}
		return nil, ErrLoggedInAsNormalUser
	}

	graph, err := h.runSignupAnonymousUserGraph(false)
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
	req *http.Request,
	clientID string,
	refreshToken string,
) (*SignupAnonymousUserResult, error) {
	ctx := req.Context()
	client := h.OAuthClientResolver.ResolveClient(clientID)
	if client == nil {
		// "invalid_client"
		return nil, apierrors.NewInvalid("invalid client ID")
	}

	if !client.HasFullAccessScope() {
		// unauthorized_client
		return nil, apierrors.NewInvalid("this client may not use anonymous user")
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

		user, err := h.UserProvider.Get(authz.UserID, accesscontrol.RoleGreatest)
		if err != nil {
			return nil, err
		}
		if !user.IsAnonymous {
			return nil, ErrLoggedInAsNormalUser
		}

		resp := protocol.TokenResponse{}
		err = h.TokenService.IssueAccessGrant(client, scopes, authz.ID, authz.UserID,
			grant.ID, oauth.GrantSessionKindOffline, refreshTokenHash, resp)
		if err != nil {
			return nil, err
		}

		return &SignupAnonymousUserResult{
			TokenResponse: resp,
		}, nil
	}

	graph, err := h.runSignupAnonymousUserGraph(true)
	if err != nil {
		return nil, err
	}

	info := authenticationinfo.T{
		UserID:          graph.MustGetUserID(),
		AuthenticatedAt: h.Clock.NowUTC(),
	}

	authz, err := h.Authorizations.CheckAndGrant(
		client.ClientID,
		info.UserID,
		scopes,
	)
	if err != nil {
		return nil, err
	}

	dpopProof := dpop.GetDPoPProof(ctx)
	dpopJKT := ""
	if dpopProof != nil {
		dpopJKT = dpopProof.JKT
	}

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
	offlineGrant, tokenHash, err := h.TokenService.IssueOfflineGrant(client, opts, resp)
	if err != nil {
		return nil, err
	}

	err = h.TokenService.IssueAccessGrant(client, scopes, authz.ID, authz.UserID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, tokenHash, resp)
	if err != nil {
		return nil, err
	}

	return &SignupAnonymousUserResult{
		TokenResponse: resp,
	}, nil
}

func (h *AnonymousUserHandler) runSignupAnonymousUserGraph(
	suppressIDPSessionCookie bool,
) (*interaction.Graph, error) {
	var graph *interaction.Graph
	err := h.Graphs.DryRun(interaction.ContextValues{}, func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		intent := &interactionintents.IntentAuthenticate{
			Kind:                     interactionintents.IntentAuthenticateKindLogin,
			SuppressIDPSessionCookie: suppressIDPSessionCookie,
		}
		graph, err = h.Graphs.NewGraph(ctx, intent)
		if err != nil {
			return nil, err
		}

		var edges []interaction.Edge
		graph, edges, err = h.Graphs.Accept(ctx, graph, &anonymousSignupWithoutKeyInput{})
		if len(edges) != 0 {
			return nil, errors.New("interaction not completed for anonymous users")
		} else if err != nil {
			return nil, err
		}

		return graph, nil
	})

	if apierrors.IsKind(err, api.InvariantViolated) &&
		apierrors.AsAPIError(err).HasCause("AnonymousUserDisallowed") {
		// unauthorized_client
		return nil, apierrors.NewInvalid("anonymous user disallowed")
	} else if errors.Is(err, api.ErrInvalidCredentials) {
		// invalid_grant
		return nil, apierrors.NewInvalid(api.InvalidCredentials.Reason)
	} else if err != nil {
		return nil, err
	}

	err = h.Graphs.Run(interaction.ContextValues{}, graph)
	if apierrors.IsAPIError(err) {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return graph, nil
}

func (h *AnonymousUserHandler) IssuePromotionCode(
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
		authz, _, _, e := h.TokenService.ParseRefreshToken(req.Context(), refreshToken)
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
		s := session.GetSession(req.Context())
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

	user, err := h.UserProvider.Get(userID, accesscontrol.RoleGreatest)
	if err != nil {
		return
	}
	if !user.IsAnonymous {
		err = ErrLoggedInAsNormalUser
		return
	}

	identities, err := h.AnonymousIdentities.List(userID)
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
	err = h.PromotionCodes.CreatePromotionCode(cObj)
	if err != nil {
		return
	}
	code = c
	codeObj = cObj
	return
}
