package hook

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
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
		var err error
		verifyConfig := &config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{Key: "email"},
			},
		}
		loginIDProvider := &mockLoginIDProvider{}
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
			loginIDProvider.Identities = []loginid.Identity{
				{
					ID:         "principal-id-1",
					UserID:     user.ID,
					LoginIDKey: "email",
					LoginID:    "test-1@example.com",
				},
				{
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
			authInfo := authInfoStore.AuthInfoMap[user.ID]
			So(authInfoStore.AuthInfoMap, ShouldResemble, map[string]authinfo.AuthInfo{
				user.ID: authinfo.AuthInfo{
					ID:               user.ID,
					Disabled:         user.Disabled,
					Verified:         authInfo.Verified,
					ManuallyVerified: user.ManuallyVerified,
					VerifyInfo:       user.VerifyInfo,
				},
			})
			So(authInfo.IsVerified(), ShouldEqual, user.Verified)
			So(userProfileStore.Data, ShouldResemble, map[string]map[string]interface{}{
				user.ID: user.Metadata,
			})
		}

		newBool := func(v bool) *bool { return &v }

		mutator := NewMutator(verifyConfig, loginIDProvider, authInfoStore, userProfileStore)

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

		Convey("should mutate manually verified status & auto update verified status", func() {
			user := model.User{
				ID:               "user-id",
				ManuallyVerified: false,
				Verified:         false,
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
				IsManuallyVerified: newBool(true),
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:               "user-id",
				ManuallyVerified: true,
				Verified:         true,
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

		Convey("should accumulate verification state mutations", func() {
			user := model.User{
				ID:               "user-id",
				ManuallyVerified: false,
				Verified:         false,
				VerifyInfo:       map[string]bool{},
			}
			ev := event.Event{
				Payload: event.UserUpdateEvent{
					User: user,
				},
			}
			initUser(user)
			mutator = mutator.New(&ev, &user)
			initialUser := user

			err = mutator.Add(event.Mutations{
				VerifyInfo: &map[string]bool{
					"test-1@example.com": true,
					"test-2@example.com": true,
				},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:               "user-id",
				ManuallyVerified: false,
				Verified:         true,
				VerifyInfo: map[string]bool{
					"test-1@example.com": true,
					"test-2@example.com": true,
				},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserUpdateEvent{
					User:       initialUser,
					IsVerified: newBool(true),
					VerifyInfo: &map[string]bool{
						"test-1@example.com": true,
						"test-2@example.com": true,
					},
				},
			})

			err = mutator.Add(event.Mutations{
				VerifyInfo: &map[string]bool{},
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:               "user-id",
				ManuallyVerified: false,
				Verified:         false,
				VerifyInfo:       map[string]bool{},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserUpdateEvent{
					User:       initialUser,
					IsVerified: newBool(false),
					VerifyInfo: &map[string]bool{},
				},
			})

			err = mutator.Add(event.Mutations{
				IsManuallyVerified: newBool(true),
			})
			So(err, ShouldBeNil)
			So(user, ShouldResemble, model.User{
				ID:               "user-id",
				ManuallyVerified: true,
				Verified:         true,
				VerifyInfo:       map[string]bool{},
			})
			So(ev, ShouldResemble, event.Event{
				Payload: event.UserUpdateEvent{
					User:       initialUser,
					IsVerified: newBool(true),
					VerifyInfo: &map[string]bool{},
				},
			})

			err = mutator.Apply()
			So(err, ShouldBeNil)
			testStoreData(user)
		})
	})
}
