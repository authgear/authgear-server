package declarative

import (
	"context"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

// fakeOfflineGrantSessionWithoutScope builds a real *oauth.OfflineGrantSession
// (not a mock) because oauth.SessionScopes type-asserts session.ResolvedSession
// to *oauth.OfflineGrantSession for the session.TypeOfflineGrant case — a fake
// implementation of the interface that merely reports that SessionType would
// panic there.
func fakeOfflineGrantSessionWithoutScope(userID string) *oauth.OfflineGrantSession {
	return &oauth.OfflineGrantSession{
		OfflineGrant: &oauth.OfflineGrant{
			ID:    "offline-grant-id",
			Attrs: session.Attrs{UserID: userID},
		},
		Scopes: []string{oauth.FullAccessScope},
	}
}

type fakeResolvedSessionForSelectAccount struct {
	UserID string
}

var _ session.ResolvedSession = &fakeResolvedSessionForSelectAccount{}

func (s *fakeResolvedSessionForSelectAccount) SessionID() string { return "session-id" }
func (s *fakeResolvedSessionForSelectAccount) SessionType() session.Type {
	return session.TypeIdentityProvider
}
func (s *fakeResolvedSessionForSelectAccount) SSOGroupIDPSessionID() string { return "" }
func (s *fakeResolvedSessionForSelectAccount) Session()                     {}
func (s *fakeResolvedSessionForSelectAccount) GetCreatedAt() time.Time      { return time.Time{} }
func (s *fakeResolvedSessionForSelectAccount) GetExpireAt() time.Time       { return time.Time{} }
func (s *fakeResolvedSessionForSelectAccount) GetAccessInfo() *access.Info  { return &access.Info{} }
func (s *fakeResolvedSessionForSelectAccount) GetAuthenticationInfo() authenticationinfo.T {
	return authenticationinfo.T{UserID: s.UserID}
}
func (s *fakeResolvedSessionForSelectAccount) CreateNewAuthenticationInfoByThisSession() authenticationinfo.T {
	return s.GetAuthenticationInfo()
}

type fakeIdentityServiceForSelectAccount struct {
	identitiesByUser map[string][]*identity.Info
}

var _ authflow.IdentityService = &fakeIdentityServiceForSelectAccount{}

func (s *fakeIdentityServiceForSelectAccount) New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) Get(ctx context.Context, id string) (*identity.Info, error) {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) SearchBySpec(ctx context.Context, spec *identity.Spec) (*identity.Info, []*identity.Info, error) {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error) {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) ListByUser(ctx context.Context, userID string) ([]*identity.Info, error) {
	return s.identitiesByUser[userID], nil
}
func (s *fakeIdentityServiceForSelectAccount) CheckDuplicatedByUniqueKey(ctx context.Context, info *identity.Info) (*identity.Info, error) {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) Create(ctx context.Context, is *identity.Info) error {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) Update(ctx context.Context, oldIs *identity.Info, newIs *identity.Info) error {
	panic("not implemented")
}
func (s *fakeIdentityServiceForSelectAccount) Delete(ctx context.Context, is *identity.Info) error {
	panic("not implemented")
}

func TestNewIdentificationOptionsSelectAccount(t *testing.T) {
	Convey("NewIdentificationOptionsSelectAccount", t, func() {
		makeCtx := func(resolvedSession session.ResolvedSession, sessionOpts *authflow.Session, deps *authflow.Dependencies) context.Context {
			ctx := context.Background()
			if resolvedSession != nil {
				ctx = session.WithSession(ctx, resolvedSession)
			}
			if sessionOpts == nil {
				sessionOpts = &authflow.Session{FlowID: "flow-1"}
			}
			ctx = sessionOpts.MakeContext(ctx, deps)
			return ctx
		}

		identities := &fakeIdentityServiceForSelectAccount{
			identitiesByUser: map[string][]*identity.Info{
				"user-1": {
					{
						Type: model.IdentityTypeLoginID,
						LoginID: &identity.LoginID{
							OriginalLoginID: "user@example.com",
						},
					},
				},
			},
		}
		deps := &authflow.Dependencies{
			Config:      &config.AppConfig{},
			Identities:  identities,
			HTTPRequest: &http.Request{Header: http.Header{}},
		}

		Convey("no session: option omitted", func() {
			ctx := makeCtx(nil, nil, deps)
			options, err := NewIdentificationOptionsSelectAccount(ctx, deps, authflow.Flows{}, nil, nil)
			So(err, ShouldBeNil)
			So(options, ShouldBeEmpty)
		})

		Convey("offline grant session without pre-authenticated-url scope: option omitted", func() {
			ctx := makeCtx(
				fakeOfflineGrantSessionWithoutScope("user-1"),
				&authflow.Session{FlowID: "flow-1"},
				deps,
			)
			options, err := NewIdentificationOptionsSelectAccount(ctx, deps, authflow.Flows{}, nil, nil)
			So(err, ShouldBeNil)
			So(options, ShouldBeEmpty)
		})

		Convey("suppress_idp_session_cookie: option omitted", func() {
			ctx := makeCtx(
				&fakeResolvedSessionForSelectAccount{UserID: "user-1"},
				&authflow.Session{FlowID: "flow-1", SuppressIDPSessionCookie: true},
				deps,
			)
			options, err := NewIdentificationOptionsSelectAccount(ctx, deps, authflow.Flows{}, nil, nil)
			So(err, ShouldBeNil)
			So(options, ShouldBeEmpty)
		})

		Convey("prompt=login: option omitted", func() {
			ctx := makeCtx(
				&fakeResolvedSessionForSelectAccount{UserID: "user-1"},
				&authflow.Session{FlowID: "flow-1", Prompt: []string{"login"}},
				deps,
			)
			options, err := NewIdentificationOptionsSelectAccount(ctx, deps, authflow.Flows{}, nil, nil)
			So(err, ShouldBeNil)
			So(options, ShouldBeEmpty)
		})

		Convey("usable session: option offered with display name and recorded user ID", func() {
			ctx := makeCtx(
				&fakeResolvedSessionForSelectAccount{UserID: "user-1"},
				&authflow.Session{FlowID: "flow-1"},
				deps,
			)
			options, err := NewIdentificationOptionsSelectAccount(ctx, deps, authflow.Flows{}, nil, nil)
			So(err, ShouldBeNil)
			So(options, ShouldHaveLength, 1)
			So(options[0].Option.UserID, ShouldEqual, "user-1")
			So(options[0].Option.Identification, ShouldEqual, model.AuthenticationFlowIdentificationSelectAccount)
			So(options[0].Option.DisplayName, ShouldEqual, "user@example.com")
		})

		Convey("usable session with mismatched login_hint/id_token_hint: option still offered (deferred, not implemented yet)", func() {
			ctx := makeCtx(
				&fakeResolvedSessionForSelectAccount{UserID: "user-1"},
				&authflow.Session{
					FlowID:     "flow-1",
					LoginHint:  "https://authgear.com/login_hint?type=login_id&login_id=someone-else%40example.com",
					UserIDHint: "user-2",
				},
				deps,
			)
			options, err := NewIdentificationOptionsSelectAccount(ctx, deps, authflow.Flows{}, nil, nil)
			So(err, ShouldBeNil)
			So(options, ShouldHaveLength, 1)
			So(options[0].Option.UserID, ShouldEqual, "user-1")
		})
	})
}
