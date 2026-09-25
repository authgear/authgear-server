package authflowv2

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type fakeSettingsAuthorizations struct {
	byID map[string]*oauth.Authorization
	err  error
}

func (f *fakeSettingsAuthorizations) GetByID(ctx context.Context, id string) (*oauth.Authorization, error) {
	if f.err != nil {
		return nil, f.err
	}
	if a, ok := f.byID[id]; ok {
		return a, nil
	}
	return nil, oauth.ErrAuthorizationNotFound
}

func (f *fakeSettingsAuthorizations) ListByUser(ctx context.Context, userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error) {
	panic("not used")
}

func (f *fakeSettingsAuthorizations) Delete(ctx context.Context, a *oauth.Authorization) error {
	panic("not used")
}

type fakeSettingsClientResolver map[string]*config.OAuthClientConfig

func (f fakeSettingsClientResolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	return f[clientID]
}

func isNotFound(err error) bool {
	return apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
		return e.Name == apierrors.NotFound
	})
}

func TestGetAuthorizationForUser(t *testing.T) {
	Convey("getAuthorizationForUser", t, func() {
		ctx := context.Background()
		authzs := &fakeSettingsAuthorizations{byID: map[string]*oauth.Authorization{
			"mine":         {ID: "mine", UserID: "alice", ClientID: "third-party"},
			"others":       {ID: "others", UserID: "bob", ClientID: "third-party"},
			"first-party":  {ID: "first-party", UserID: "alice", ClientID: "first-party"},
			"unresolvable": {ID: "unresolvable", UserID: "alice", ClientID: "deleted"},
		}}
		h := &AuthflowV2SettingsAuthorizedAppsHandler{
			Authorizations: authzs,
			OAuthClientResolver: fakeSettingsClientResolver{
				"third-party": {ClientID: "third-party", ApplicationType: config.OAuthClientApplicationTypeThirdPartyApp},
				"first-party": {ClientID: "first-party", ApplicationType: config.OAuthClientApplicationTypeSPA},
			},
		}

		Convey("returns the signed-in user's own third-party grant", func() {
			authz, err := h.getAuthorizationForUser(ctx, "alice", "mine")
			So(err, ShouldBeNil)
			So(authz.ID, ShouldEqual, "mine")
		})

		Convey("reports another user's grant as not found", func() {
			authz, err := h.getAuthorizationForUser(ctx, "alice", "others")
			So(authz, ShouldBeNil)
			So(isNotFound(err), ShouldBeTrue)
		})

		Convey("reports a first-party grant as not found", func() {
			authz, err := h.getAuthorizationForUser(ctx, "alice", "first-party")
			So(authz, ShouldBeNil)
			So(isNotFound(err), ShouldBeTrue)
		})

		Convey("reports a grant whose client no longer resolves as not found", func() {
			authz, err := h.getAuthorizationForUser(ctx, "alice", "unresolvable")
			So(authz, ShouldBeNil)
			So(isNotFound(err), ShouldBeTrue)
		})

		Convey("reports a missing or unknown id as not found", func() {
			_, err := h.getAuthorizationForUser(ctx, "alice", "")
			So(isNotFound(err), ShouldBeTrue)
			_, err = h.getAuthorizationForUser(ctx, "alice", "no-such-id")
			So(isNotFound(err), ShouldBeTrue)
		})

		Convey("passes any other error through rather than reporting not found", func() {
			dbErr := errors.New("connection lost")
			authzs.err = dbErr
			_, err := h.getAuthorizationForUser(ctx, "alice", "mine")
			So(err, ShouldEqual, dbErr)
			So(isNotFound(err), ShouldBeFalse)
		})
	})
}

func TestResourcePermissions(t *testing.T) {
	Convey("resourcePermissions", t, func() {
		authz := &oauth.Authorization{ScopesByResources: map[string][]string{
			oauth.AuthorizationProjectScopesKey: {"openid", "profile"},
			"r1":                                {"read:tools", "execute:tools"},
			"r2":                                {"read:tools"},
			"deleted":                           {"read:tools"},
		}}
		existing := map[string]struct{}{"r1": {}, "r2": {}}

		Convey("uses the description of the scope on the resource it is granted on", func() {
			got := resourcePermissions(authz, existing, map[resourceScopeKey]string{
				{ResourceID: "r1", Scope: "read:tools"}: "Show all tools",
				{ResourceID: "r2", Scope: "read:tools"}: "Show inventory tools",
			})
			So(got, ShouldResemble, []AuthorizationPermission{
				{Scope: "read:tools", DisplayText: "Show all tools"},
				{Scope: "execute:tools", DisplayText: "execute:tools"},
				{Scope: "read:tools", DisplayText: "Show inventory tools"},
			})
		})

		Convey("skips project-level scopes and deleted resources, and tolerates nil maps", func() {
			got := resourcePermissions(authz, nil, nil)
			So(got, ShouldBeEmpty)
		})
	})
}
