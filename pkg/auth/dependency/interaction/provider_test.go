package interaction_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/clock"
)

func TestInteractionProviderProgrammingError(t *testing.T) {
	Convey("InteractionProviderProgrammingError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identityProvider := NewMockIdentityProvider(ctrl)
		authenticatorProvider := NewMockAuthenticatorProvider(ctrl)
		store := NewMockStore(ctrl)

		p := &interaction.Provider{
			Clock:         clock.NewMockClock(),
			Identity:      identityProvider,
			Authenticator: authenticatorProvider,
			Store:         store,
		}
		i := &interaction.Interaction{
			Intent:   &interaction.IntentLogin{},
			Identity: &identity.Ref{},
		}
		identityInfo := &identity.Info{}

		store.EXPECT().Create(gomock.Any()).Return(nil).AnyTimes()
		store.EXPECT().Delete(gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().CreateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(identityInfo, nil).AnyTimes()
		authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		authenticatorProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		authenticatorProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		Convey("panic if commit after save", func() {
			_, err := p.SaveInteraction(i)
			So(err, ShouldBeNil)
			So(func() { p.Commit(i) }, ShouldPanic)
		})

		Convey("panic if save after commit", func() {
			_, err := p.Commit(i)
			So(err, ShouldBeNil)
			So(func() { p.SaveInteraction(i) }, ShouldPanic)
		})
	})
}
