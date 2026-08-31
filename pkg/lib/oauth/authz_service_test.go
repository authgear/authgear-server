package oauth_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type stubAuthorizationStore struct {
	authzs []*oauth.Authorization
}

func (s *stubAuthorizationStore) Get(ctx context.Context, userID, clientID string) (*oauth.Authorization, error) {
	panic("not used by this test")
}
func (s *stubAuthorizationStore) GetByID(ctx context.Context, id string) (*oauth.Authorization, error) {
	panic("not used by this test")
}
func (s *stubAuthorizationStore) ListByUserID(ctx context.Context, userID string) ([]*oauth.Authorization, error) {
	return s.authzs, nil
}
func (s *stubAuthorizationStore) Create(ctx context.Context, a *oauth.Authorization) error {
	panic("not used by this test")
}
func (s *stubAuthorizationStore) Delete(ctx context.Context, a *oauth.Authorization) error {
	panic("not used by this test")
}
func (s *stubAuthorizationStore) ResetAll(ctx context.Context, userID string) error {
	panic("not used by this test")
}
func (s *stubAuthorizationStore) UpdateScopes(ctx context.Context, a *oauth.Authorization) error {
	panic("not used by this test")
}

var _ oauth.AuthorizationStore = &stubAuthorizationStore{}

func TestAuthorizationServiceListByUser(t *testing.T) {
	Convey("AuthorizationService.ListByUser", t, func() {
		ctx := context.Background()
		store := &stubAuthorizationStore{
			authzs: []*oauth.Authorization{
				{ID: "kept", ClientID: "kept-client"},
				{ID: "dropped-by-first", ClientID: "dropped-by-first-client"},
			},
		}
		svc := &oauth.AuthorizationService{Store: store}

		Convey("when the first filter drops an authorization, the second filter is not consulted for it", func() {
			var secondFilterCalledFor []string
			first := oauth.AuthorizationFilterFunc(func(ctx context.Context, a *oauth.Authorization) bool {
				return a.ClientID != "dropped-by-first-client"
			})
			second := oauth.AuthorizationFilterFunc(func(ctx context.Context, a *oauth.Authorization) bool {
				secondFilterCalledFor = append(secondFilterCalledFor, a.ClientID)
				return true
			})

			result, err := svc.ListByUser(ctx, "user-1", first, second)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].ClientID, ShouldEqual, "kept-client")
			// The second filter must only ever have been asked about the
			// authorization the first filter kept -- the short-circuit
			// (break on the first false) must be preserved.
			So(secondFilterCalledFor, ShouldResemble, []string{"kept-client"})
		})
	})
}
