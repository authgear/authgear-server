package interaction_test

import (
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestProviderCommit(t *testing.T) {
	Convey("InteractionProviderCommitRemoveIdentity", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identityProvider := NewMockIdentityProvider(ctrl)
		authenticatorProvider := NewMockAuthenticatorProvider(ctrl)
		userProvider := NewMockUserProvider(ctrl)
		hooks := NewMockHookProvider(ctrl)

		p := &interaction.Provider{
			Clock:         clock.NewMockClock(),
			Identity:      identityProvider,
			Authenticator: authenticatorProvider,
			User:          userProvider,
			Hooks:         hooks,
		}
		userID := "userid1"
		loginID1 := &identity.Info{
			ID:     "iid1",
			Type:   authn.IdentityTypeLoginID,
			Claims: make(map[string]interface{}),
		}
		loginID2 := &identity.Info{
			ID:     "iid2",
			Type:   authn.IdentityTypeLoginID,
			Claims: make(map[string]interface{}),
		}
		oauthID := &identity.Info{
			ID:     "iid3",
			Type:   authn.IdentityTypeOAuth,
			Claims: make(map[string]interface{}),
		}
		pwAuthenticator := &authenticator.Info{
			ID:   "aid1",
			Type: authn.AuthenticatorTypePassword,
		}
		totpAuthenticator := &authenticator.Info{
			ID:   "aid2",
			Type: authn.AuthenticatorTypeTOTP,
		}
		oobAuthenticator := &authenticator.Info{
			ID:   "aid3",
			Type: authn.AuthenticatorTypeOOB,
		}

		authenticatorProvider.EXPECT().ListByIdentity(userID, loginID1).Return([]*authenticator.Info{
			pwAuthenticator, totpAuthenticator,
		}, nil).AnyTimes()
		authenticatorProvider.EXPECT().ListByIdentity(userID, loginID2).Return([]*authenticator.Info{
			pwAuthenticator, totpAuthenticator, oobAuthenticator,
		}, nil).AnyTimes()
		authenticatorProvider.EXPECT().ListByIdentity(userID, oauthID).Return([]*authenticator.Info{}, nil).AnyTimes()

		identityProvider.EXPECT().CreateAll(gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().UpdateAll(gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().DeleteAll(gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity.Info{}, nil).AnyTimes()
		authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		authenticatorProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		authenticatorProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		userProvider.EXPECT().Get(userID).Return(&model.User{ID: userID}, nil).AnyTimes()

		Convey("should cleanup authenticators", func() {
			// remove login id
			i := &interaction.Interaction{
				Intent:           &interaction.IntentRemoveIdentity{},
				Identity:         &identity.Ref{},
				UserID:           userID,
				RemoveIdentities: []*identity.Info{loginID1},
			}
			// user has 1 login id and 1 oauth identity
			identityProvider.EXPECT().ListByUser(gomock.Any()).Return([]*identity.Info{loginID1, oauthID}, nil).AnyTimes()

			hooks.EXPECT().DispatchEvent(event.IdentityDeleteEvent{
				User: model.User{ID: userID},
				Identity: model.Identity{
					Type:   string(loginID1.Type),
					Claims: loginID1.Claims,
				},
			}, gomock.Any())

			_, err := p.Commit(i)
			So(err, ShouldBeNil)

			expected := i.RemoveAuthenticators
			actual := []*authenticator.Info{
				pwAuthenticator, totpAuthenticator,
			}
			sort.Sort(authenticatorInfoSlice(expected))
			sort.Sort(authenticatorInfoSlice(actual))
			So(expected, ShouldResemble, actual)
		})

		Convey("should not remove authenticators when removing identity has no related authenticator", func() {
			// remove oauth identity
			i := &interaction.Interaction{
				Intent:           &interaction.IntentRemoveIdentity{},
				Identity:         &identity.Ref{},
				UserID:           userID,
				RemoveIdentities: []*identity.Info{oauthID},
			}
			// user has 1 login id and 1 oauth identity
			identityProvider.EXPECT().ListByUser(gomock.Any()).Return([]*identity.Info{loginID1, oauthID}, nil).AnyTimes()

			hooks.EXPECT().DispatchEvent(event.IdentityDeleteEvent{
				User: model.User{ID: userID},
				Identity: model.Identity{
					Type:   string(oauthID.Type),
					Claims: oauthID.Claims,
				},
			}, gomock.Any())

			_, err := p.Commit(i)
			So(err, ShouldBeNil)

			So(len(i.RemoveAuthenticators), ShouldEqual, 0)

		})

		Convey("should keep authenticators which related to existing identities", func() {
			// remove oauth identity
			i := &interaction.Interaction{
				Intent:           &interaction.IntentRemoveIdentity{},
				Identity:         &identity.Ref{},
				UserID:           userID,
				RemoveIdentities: []*identity.Info{loginID2},
			}
			// user has 2 login id and 1 oauth identity
			identityProvider.EXPECT().ListByUser(gomock.Any()).Return([]*identity.Info{loginID1, loginID2, oauthID}, nil).AnyTimes()

			hooks.EXPECT().DispatchEvent(event.IdentityDeleteEvent{
				User: model.User{ID: userID},
				Identity: model.Identity{
					Type:   string(loginID2.Type),
					Claims: loginID2.Claims,
				},
			}, gomock.Any())

			_, err := p.Commit(i)
			So(err, ShouldBeNil)

			// pw and otp authenticators are used by login id 1 which should be kept

			expected := i.RemoveAuthenticators
			actual := []*authenticator.Info{
				oobAuthenticator,
			}
			sort.Sort(authenticatorInfoSlice(expected))
			sort.Sort(authenticatorInfoSlice(actual))
			So(expected, ShouldResemble, actual)
		})
	})
}

type authenticatorInfoSlice []*authenticator.Info

func (s authenticatorInfoSlice) Len() int           { return len(s) }
func (s authenticatorInfoSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s authenticatorInfoSlice) Less(i, j int) bool { return s[i].ID < s[j].ID }
