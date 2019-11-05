package hook

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMutator(t *testing.T) {
	Convey("Hook Mutator", t, func() {
		var err error
		verifyConfig := config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{Key: "email"},
			},
		}
		passwordAuthProvider := password.NewMockProvider(
			[]config.LoginIDKeyConfiguration{},
			[]string{},
		)
		authInfoStore := authinfo.NewMockStore()
		userProfileStore := userprofile.NewMockUserProfileStore()

		initUser := func(user model.User) {
			authInfoStore.AuthInfoMap = map[string]authinfo.AuthInfo{
				user.ID: authinfo.AuthInfo{
					ID:         user.ID,
					Disabled:   user.Disabled,
					Verified:   user.Verified,
					VerifyInfo: user.VerifyInfo,
				},
			}
			passwordAuthProvider.PrincipalMap = map[string]password.Principal{
				"principal-id-1": password.Principal{
					ID:         "principal-id-1",
					UserID:     user.ID,
					LoginIDKey: "email",
					LoginID:    "test-1@example.com",
				},
				"principal-id-2": password.Principal{
					ID:         "principal-id-2",
					UserID:     user.ID,
					LoginIDKey: "email",
					LoginID:    "test-2@example.com",
				},
			}
			userProfileStore.Data = map[string]map[string]interface{}{
				user.ID: user.Metadata,
			}
		}
		testStoreData := func(user model.User) {
			So(authInfoStore.AuthInfoMap, ShouldResemble, map[string]authinfo.AuthInfo{
				user.ID: authinfo.AuthInfo{
					ID:         user.ID,
					Disabled:   user.Disabled,
					Verified:   user.Verified,
					VerifyInfo: user.VerifyInfo,
				},
			})
			So(userProfileStore.Data, ShouldResemble, map[string]map[string]interface{}{
				user.ID: user.Metadata,
			})
		}

		newBool := func(v bool) *bool { return &v }

		mutator := NewMutator(&verifyConfig, passwordAuthProvider, authInfoStore, userProfileStore)

		Convey("should do nothing", func() {
			user := model.User{
				ID: "user-id",
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			initUser(user)
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
			testStoreData(user)
		})

		Convey("should mutate metadata only", func() {
			user := model.User{
				ID:       "user-id",
				Verified: true,
				Metadata: userprofile.Data{
					"test": 123,
				},
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			initUser(user)
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				Metadata: &userprofile.Data{
					"example": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:       "user-id",
				Verified: true,
				Metadata: userprofile.Data{
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
			testStoreData(user)
		})

		Convey("should mutate verified status", func() {
			user := model.User{
				ID:       "user-id",
				Verified: false,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
				},
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			initUser(user)
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				IsVerified: newBool(true),
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:       "user-id",
				Verified: true,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
			testStoreData(user)
		})

		Convey("should mutate verify info & auto update verified status", func() {
			user := model.User{
				ID:       "user-id",
				Verified: false,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
				},
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			initUser(user)
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				VerifyInfo: &map[string]bool{
					"test-1@example.com": true,
					"test-2@example.com": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:       "user-id",
				Verified: true,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
					"test-2@example.com": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
			testStoreData(user)
		})

		Convey("should allow overriding verified status", func() {
			user := model.User{
				ID:       "user-id",
				Verified: false,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
				},
			}
			ev := event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			}
			initUser(user)
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				IsVerified: newBool(false),
				VerifyInfo: &map[string]bool{
					"test-1@example.com": true,
					"test-2@example.com": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:       "user-id",
				Verified: false,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
					"test-2@example.com": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
			testStoreData(user)
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
			initUser(user)
			mutator = mutator.New(&ev, &user)

			err = mutator.Add(event.Mutations{
				Metadata: &userprofile.Data{
					"example": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID: "user-id",
				Metadata: userprofile.Data{
					"example": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserSyncEvent{
					User: user,
				},
			})

			err = mutator.Add(event.Mutations{
				IsDisabled: newBool(true),
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:       "user-id",
				Disabled: true,
				Metadata: userprofile.Data{
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
			testStoreData(user)
		})
	})
}
