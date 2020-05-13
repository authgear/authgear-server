package interaction_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

func TestProvider(t *testing.T) {
	SkipConvey("Interaction Provider", t, func() {
		var p *interaction.Provider

		Convey("Common password flow", func() {
			Convey("Login", func() {
				Convey("step 1", func() {
					i, err := p.NewInteractionLogin(
						&interaction.IntentLogin{Identity: identity.Spec{
							Type:   authn.IdentityTypeLoginID,
							Claims: map[string]interface{}{"email": "user@example.com"},
						}},
						"",
					)
					So(err, ShouldBeNil)

					state, err := p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.Steps, ShouldHaveLength, 1)
					So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticatePrimary)
					So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
						Type:  authn.AuthenticatorTypePassword,
						Props: map[string]interface{}{},
					})

					token, err := p.SaveInteraction(i)
					So(err, ShouldBeNil)
					So(token, ShouldNotBeEmpty)
				})
				Convey("step 2", func() {
					var token string
					i, err := p.GetInteraction(token)
					So(err, ShouldBeNil)

					state, err := p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.Steps, ShouldHaveLength, 1)
					So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticatePrimary)
					So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
						Type:  authn.AuthenticatorTypePassword,
						Props: map[string]interface{}{},
					})

					err = p.PerformAction(i, interaction.StepAuthenticatePrimary, &interaction.ActionAuthenticate{
						Authenticator: state.Steps[0].AvailableAuthenticators[0],
						Secret:        "password",
					})
					So(err, ShouldBeNil)

					state, err = p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.Steps, ShouldHaveLength, 2)
					So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticatePrimary)
					So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

					_, err = p.Commit(i)
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("SSO flow with MFA", func() {
			Convey("step 1", func() {
				i, err := p.NewInteractionLogin(
					&interaction.IntentLogin{
						Identity: identity.Spec{
							Type: authn.IdentityTypeOAuth,
							Claims: map[string]interface{}{
								identity.IdentityClaimOAuthProvider: map[string]interface{}{
									"type":   "azureadv2",
									"tenant": "example",
								},
								identity.IdentityClaimOAuthSubjectID: "9A8822AA-4F18-4E4C-84AF-E0FD9AB86CB2",
								identity.IdentityClaimOAuthProfile:   map[string]interface{}{},
							},
						},
					},
					"",
				)
				So(err, ShouldBeNil)

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticateSecondary)
				So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
					Type: authn.AuthenticatorTypeTOTP,
					Props: map[string]interface{}{
						authenticator.AuthenticatorPropTOTPDisplayName: "My Authenticator",
					},
				})

				token, err := p.SaveInteraction(i)
				So(err, ShouldBeNil)
				So(token, ShouldNotBeEmpty)
			})
			Convey("step 2", func() {
				var token string
				i, err := p.GetInteraction(token)
				So(err, ShouldBeNil)

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticateSecondary)
				So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
					Type: authn.AuthenticatorTypeTOTP,
					Props: map[string]interface{}{
						authenticator.AuthenticatorPropTOTPDisplayName: "My Authenticator",
					},
				})

				err = p.PerformAction(i, interaction.StepAuthenticateSecondary, &interaction.ActionAuthenticate{
					Authenticator: state.Steps[0].AvailableAuthenticators[0],
					Secret:        "123456",
				})
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 2)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticateSecondary)
				So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

				_, err = p.Commit(i)
				So(err, ShouldBeNil)
			})
		})

		Convey("Setup MFA", func() {
			Convey("step 1", func() {
				i, err := p.NewInteractionAddAuthenticator(
					&interaction.IntentAddAuthenticator{
						Authenticator: authenticator.Spec{
							Type: authn.AuthenticatorTypeTOTP,
							Props: map[string]interface{}{
								authenticator.AuthenticatorPropTOTPDisplayName: "My Authenticator",
							},
						},
					},
					"",
					nil,
				)
				So(err, ShouldBeNil)

				So(i.NewAuthenticators, ShouldNotBeEmpty)
				So(i.NewAuthenticators, ShouldResemble, []authenticator.Spec{
					{
						Type: authn.AuthenticatorTypeTOTP,
						Props: map[string]interface{}{
							authenticator.AuthenticatorPropTOTPDisplayName: "My Authenticator",
						},
					},
				})

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupSecondaryAuthenticator)

				err = p.PerformAction(i, interaction.StepSetupSecondaryAuthenticator, &interaction.ActionAuthenticate{
					Secret: "123456",
				})
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 2)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupSecondaryAuthenticator)
				So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

				_, err = p.Commit(i)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestInteractionProviderProgrammingError(t *testing.T) {
	Convey("InteractionProviderProgrammingError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identityProvider := NewMockIdentityProvider(ctrl)
		authenticatorProvider := NewMockAuthenticatorProvider(ctrl)
		store := NewMockStore(ctrl)

		p := &interaction.Provider{
			Time:          &coretime.MockProvider{},
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
