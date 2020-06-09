package hook

import (
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

type mockLoginIDProvider struct {
	Identities []loginid.Identity
}

func (p *mockLoginIDProvider) List(userID string) ([]*loginid.Identity, error) {
	var is []*loginid.Identity
	for _, i := range p.Identities {
		if i.UserID == userID {
			ii := i
			is = append(is, &ii)
		}
	}
	return is, nil
}

func TestMutator(t *testing.T) {
	Convey("Hook Mutator", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var err error
		verifyConfig := &config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{Key: "email"},
			},
		}
		loginIDProvider := &mockLoginIDProvider{}
		loginIDProvider.Identities = []loginid.Identity{
			{
				ID:         "principal-id-1",
				UserID:     "user-id",
				LoginIDKey: "email",
				LoginID:    "test-1@example.com",
			},
			{
				ID:         "principal-id-2",
				UserID:     "user-id",
				LoginIDKey: "email",
				LoginID:    "test-2@example.com",
			},
		}

		users := NewMockUserProvider(ctrl)

		mutator := NewMutator(verifyConfig, loginIDProvider, users)

		Convey("should do nothing", func() {
			user := model.User{
				ID: "user-id",
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{})
			So(err, ShouldBeNil)

			So(user, ShouldResemble, model.User{
				ID: "user-id",
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
		})

		Convey("should mutate metadata only", func() {
			user := model.User{
				ID: "user-id",
				Metadata: map[string]interface{}{
					"test": 123,
				},
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			users.EXPECT().UpdateMetadata(&user, map[string]interface{}{
				"example": true,
			}).Return(nil)
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				Metadata: &map[string]interface{}{
					"example": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID: "user-id",
				Metadata: map[string]interface{}{
					"example": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
		})

		Convey("should accumulate mutations", func() {
			user := model.User{
				ID: "user-id",
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			users.EXPECT().UpdateMetadata(&user, map[string]interface{}{
				"example-1": false,
			}).Do(func(user *model.User, metadata map[string]interface{}) error {
				user.Metadata = metadata
				return nil
			})

			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				Metadata: &map[string]interface{}{
					"example": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID: "user-id",
				Metadata: map[string]interface{}{
					"example": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Add(event.Mutations{
				Metadata: &map[string]interface{}{
					"example-1": false,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID: "user-id",
				Metadata: map[string]interface{}{
					"example-1": false,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
		})
	})
}
