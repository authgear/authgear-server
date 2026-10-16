package handler

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type fakePreAuthenticatedURLTokenStore struct {
	token *oauth.PreAuthenticatedURLToken
}

func (s *fakePreAuthenticatedURLTokenStore) CreatePreAuthenticatedURLToken(ctx context.Context, t *oauth.PreAuthenticatedURLToken) error {
	return nil
}

func (s *fakePreAuthenticatedURLTokenStore) ConsumePreAuthenticatedURLToken(ctx context.Context, tokenHash string) (*oauth.PreAuthenticatedURLToken, error) {
	return s.token, nil
}

type fakePreAuthenticatedURLAccessGrantService struct {
	options oauth.PrepareUserAccessGrantOptions
}

func (s *fakePreAuthenticatedURLAccessGrantService) PrepareUserAccessGrant(ctx context.Context, options oauth.PrepareUserAccessGrantOptions) (oauth.PrepareUserAccessTokenResult, error) {
	s.options = options
	return nil, nil
}

type fakePreAuthenticatedURLOfflineGrantService struct {
	grant   *oauth.OfflineGrant
	created oauth.CreateNewRefreshTokenOptions
}

func (s *fakePreAuthenticatedURLOfflineGrantService) GetOfflineGrant(ctx context.Context, id string) (*oauth.OfflineGrant, error) {
	return s.grant, nil
}

func (s *fakePreAuthenticatedURLOfflineGrantService) CreateNewRefreshToken(ctx context.Context, options oauth.CreateNewRefreshTokenOptions) (*oauth.CreateNewRefreshTokenResult, *oauth.OfflineGrant, error) {
	s.created = options
	return &oauth.CreateNewRefreshTokenResult{TokenHash: "new-token-hash"}, s.grant, nil
}

// The browser session of a session that was never bound to a resource must
// carry exactly the scopes the pre-authenticated URL token was made with.
// The service is built without a resource access policy on purpose: an
// unbound token has no resource to look up.
func TestPreAuthenticatedURLTokenServiceExchangeForAccessToken(t *testing.T) {
	Convey("PreAuthenticatedURLTokenServiceImpl.ExchangeForAccessToken", t, func() {
		ctx := context.Background()
		client := &config.OAuthClientConfig{
			ClientID:        "web-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		scopes := []string{
			"openid",
			"offline_access",
			"device_sso",
			"https://authgear.com/scopes/pre-authenticated-url",
			"https://authgear.com/scopes/full-access",
		}
		grant := &oauth.OfflineGrant{ID: "offline-grant-id"}

		store := &fakePreAuthenticatedURLTokenStore{token: &oauth.PreAuthenticatedURLToken{
			ClientID:        client.ClientID,
			OfflineGrantID:  grant.ID,
			AuthorizationID: "authz-id",
			Scopes:          scopes,
		}}
		accessGrants := &fakePreAuthenticatedURLAccessGrantService{}
		offlineGrants := &fakePreAuthenticatedURLOfflineGrantService{grant: grant}

		s := &PreAuthenticatedURLTokenServiceImpl{
			Clock:                     clock.NewMockClockAt("2026-01-01T00:00:00Z"),
			PreAuthenticatedURLTokens: store,
			AccessGrantService:        accessGrants,
			OfflineGrantService:       offlineGrants,
		}

		_, err := s.ExchangeForAccessToken(ctx, client, grant.ID, "token")
		So(err, ShouldBeNil)

		So(offlineGrants.created.ClientID, ShouldEqual, client.ClientID)
		So(offlineGrants.created.Scopes, ShouldResemble, scopes)
		So(accessGrants.options.Scopes, ShouldResemble, scopes)
		So(accessGrants.options.AuthorizationID, ShouldEqual, "authz-id")
	})
}
