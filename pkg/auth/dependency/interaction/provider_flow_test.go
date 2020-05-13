package interaction_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

func TestProviderFlow(t *testing.T) {
	Convey("Interaction Provider", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identityProvider := NewMockIdentityProvider(ctrl)
		authenticatorProvider := NewMockAuthenticatorProvider(ctrl)
		store := NewMockStore(ctrl)
		userProvider := NewMockUserProvider(ctrl)
		hooks := hook.NewMockProvider()

		p := &interaction.Provider{
			Time:          &coretime.MockProvider{},
			Identity:      identityProvider,
			Authenticator: authenticatorProvider,
			User:          userProvider,
			Hooks:         hooks,
			Store:         store,
		}

		Convey("Common password flow", func() {
			authnConfig := &config.AuthenticationConfiguration{
				PrimaryAuthenticators: []string{"password"},
			}

			p.Config = authnConfig

			Convey("Signup", func() {

				// step 1 setup
				loginIDClaims := map[string]interface{}{"email": "user@example.com"}
				is := identity.Spec{
					Type:   authn.IdentityTypeLoginID,
					Claims: loginIDClaims,
				}
				ii := &identity.Info{
					ID:     "identity_id_1",
					Type:   authn.IdentityTypeLoginID,
					Claims: loginIDClaims,
				}
				as := &authenticator.Spec{
					Type:  authn.AuthenticatorTypePassword,
					Props: map[string]interface{}{},
				}
				ai := &authenticator.Info{
					ID:     "authenticator_id_1",
					Type:   authn.AuthenticatorTypePassword,
					Props:  map[string]interface{}{},
					Secret: "password",
				}
				identityProvider.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Eq(loginIDClaims)).Return(ii)
				identityProvider.EXPECT().Validate(gomock.Any()).Return(nil)
				identityProvider.EXPECT().RelateIdentityToAuthenticator(gomock.Eq(is), gomock.Eq(as)).Return(as).AnyTimes()
				store.EXPECT().Create(gomock.Any()).Return(nil)

				// step 1
				i, err := p.NewInteractionSignup(
					&interaction.IntentSignup{
						Identity: identity.Spec{
							Type:   authn.IdentityTypeLoginID,
							Claims: loginIDClaims,
						},
						UserMetadata: map[string]interface{}{},
					},
					"",
				)
				So(err, ShouldBeNil)

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupPrimaryAuthenticator)
				So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
					Type:  authn.AuthenticatorTypePassword,
					Props: map[string]interface{}{},
				})

				iCopy := *i
				token, err := p.SaveInteraction(i)
				So(err, ShouldBeNil)
				So(token, ShouldNotBeEmpty)

				// step 2 setup
				store.EXPECT().Get(gomock.Eq(token)).Return(&iCopy, nil)
				store.EXPECT().Delete(gomock.Any()).Return(nil)

				userProvider.EXPECT().Create(
					gomock.Any(), gomock.Any(), gomock.Eq([]*identity.Info{ii}),
				).Return(nil)

				identityProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq([]*identity.Info{ii})).Return(nil)
				var emptyIdentityInfoList []*identity.Info
				identityProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().Get(gomock.Any(), ii.Type, ii.ID).Return(ii, nil)

				authenticatorProvider.EXPECT().New(
					gomock.Any(), gomock.Eq(*as), gomock.Eq("password"),
				).Return([]*authenticator.Info{ai}, nil)
				authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq([]*authenticator.Info{ai})).Return(nil)
				var emptyAuthenticatorInfoList []*authenticator.Info
				authenticatorProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyAuthenticatorInfoList)).Return(nil)

				// step 2
				i2, err := p.GetInteraction(token)
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i2)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupPrimaryAuthenticator)
				So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
					Type:  authn.AuthenticatorTypePassword,
					Props: map[string]interface{}{},
				})

				err = p.PerformAction(i2, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
					Authenticator: state.Steps[0].AvailableAuthenticators[0],
					Secret:        "password",
				})
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i2)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 2)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupPrimaryAuthenticator)
				So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

				_, err = p.Commit(i2)
				So(err, ShouldBeNil)
			})
		})

	})
}
