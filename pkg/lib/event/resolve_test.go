package event

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type mockResolverUserQueriesCall struct {
	id   string
	role accesscontrol.Role
}

type mockResolverUserQueries struct {
	calls []mockResolverUserQueriesCall
	users map[string]*model.User
}

func (m *mockResolverUserQueries) Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
	m.calls = append(m.calls, mockResolverUserQueriesCall{id: id, role: role})
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

type mockResolvePayload struct {
	UserRef model.UserRef `json:"-" resolve:"user"`
	User    model.User    `json:"user"`
}

func TestResolverImplResolve(t *testing.T) {
	Convey("ResolverImpl.Resolve", t, func() {
		Convey("populates the json:\"user\" field from Users.Get", func() {
			resolvedUser := model.User{Meta: model.Meta{ID: "user-a"}}
			users := &mockResolverUserQueries{users: map[string]*model.User{"user-a": &resolvedUser}}
			r := &ResolverImpl{Users: users}

			payload := &mockResolvePayload{UserRef: model.UserRef{Meta: model.Meta{ID: "user-a"}}}
			err := r.Resolve(context.Background(), payload)

			So(err, ShouldBeNil)
			So(payload.User, ShouldResemble, resolvedUser)
			So(users.calls, ShouldResemble, []mockResolverUserQueriesCall{{id: "user-a", role: accesscontrol.RoleGreatest}})
		})

		Convey("swallows ErrUserNotFound and leaves a previously populated field intact", func() {
			users := &mockResolverUserQueries{users: map[string]*model.User{}}
			r := &ResolverImpl{Users: users}

			previouslyResolved := model.User{Meta: model.Meta{ID: "user-a"}}
			payload := &mockResolvePayload{
				UserRef: model.UserRef{Meta: model.Meta{ID: "user-a"}},
				User:    previouslyResolved,
			}
			err := r.Resolve(context.Background(), payload)

			// The walk does not overwrite the field on ErrUserNotFound; it
			// only "continue"s past that field, so the field keeps whatever
			// was there before. The error itself is still returned to the
			// caller unchanged.
			So(errors.Is(err, user.ErrUserNotFound), ShouldBeTrue)
			So(payload.User, ShouldResemble, previouslyResolved)
		})
	})
}

func TestResolverImplResolveWithUser(t *testing.T) {
	Convey("ResolverImpl.ResolveWithUser", t, func() {
		Convey("with a matching ID does not call Users.Get", func() {
			users := &mockResolverUserQueries{users: map[string]*model.User{}}
			r := &ResolverImpl{Users: users}

			override := &model.User{Meta: model.Meta{ID: "user-a"}, IsVerified: true}
			payload := &mockResolvePayload{UserRef: model.UserRef{Meta: model.Meta{ID: "user-a"}}}
			err := r.ResolveWithUser(context.Background(), payload, override)

			So(err, ShouldBeNil)
			So(payload.User, ShouldResemble, *override)
			So(users.calls, ShouldBeEmpty)
		})

		Convey("with a non-matching ID falls back to Users.Get", func() {
			resolvedUser := model.User{Meta: model.Meta{ID: "user-b"}}
			users := &mockResolverUserQueries{users: map[string]*model.User{"user-b": &resolvedUser}}
			r := &ResolverImpl{Users: users}

			override := &model.User{Meta: model.Meta{ID: "user-a"}}
			payload := &mockResolvePayload{UserRef: model.UserRef{Meta: model.Meta{ID: "user-b"}}}
			err := r.ResolveWithUser(context.Background(), payload, override)

			So(err, ShouldBeNil)
			So(payload.User, ShouldResemble, resolvedUser)
			So(users.calls, ShouldResemble, []mockResolverUserQueriesCall{{id: "user-b", role: accesscontrol.RoleGreatest}})
		})

		Convey("on UserAnonymousPromotedEventPayload uses the override for resolve:\"user\" and still calls Users.Get for resolve:\"anonymous_user\"", func() {
			anonymousUser := model.User{Meta: model.Meta{ID: "anon-user"}}
			users := &mockResolverUserQueries{users: map[string]*model.User{"anon-user": &anonymousUser}}
			r := &ResolverImpl{Users: users}

			override := &model.User{Meta: model.Meta{ID: "user-a"}, IsVerified: true}
			payload := &nonblocking.UserAnonymousPromotedEventPayload{
				AnonymousUserRef: model.UserRef{Meta: model.Meta{ID: "anon-user"}},
				UserRef:          model.UserRef{Meta: model.Meta{ID: "user-a"}},
			}
			err := r.ResolveWithUser(context.Background(), payload, override)

			So(err, ShouldBeNil)
			So(payload.UserModel, ShouldResemble, *override)
			So(payload.AnonymousUserModel, ShouldResemble, anonymousUser)
			So(users.calls, ShouldResemble, []mockResolverUserQueriesCall{{id: "anon-user", role: accesscontrol.RoleGreatest}})
		})
	})
}
