package hook

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMutator(t *testing.T) {
	Convey("Hook Mutator", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		users := NewMockUserProvider(ctrl)

		mf := &MutatorFactory{
			Users: users,
		}

		Convey("should do nothing", func() {
			user := model.User{
				ID: "user-id",
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			mutator := mf.New(&ev, &user)

			err := mutator.Add(event.Mutations{})
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
			mutator := mf.New(&ev, &user)

			err := mutator.Add(event.Mutations{
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

			mutator := mf.New(&ev, &user)

			err := mutator.Add(event.Mutations{
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
