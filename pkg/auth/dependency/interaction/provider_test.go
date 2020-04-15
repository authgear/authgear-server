package interaction_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
)

func TestProvider(t *testing.T) {
	SkipConvey("Interaction Provider", t, func() {
		var p *interaction.Provider

		Convey("Common password flow", func() {
			Convey("Signup", func() {
				Convey("step 1", func() {
					i, err := p.NewInteraction(
						interaction.IntentSignup{
							Identity: interaction.IdentitySpec{
								Type:   interaction.IdentityTypeLoginID,
								Claims: map[string]interface{}{"email": "user@example.com"},
							},
							UserMetadata: map[string]interface{}{},
						},
						nil,
					)
					So(err, ShouldBeNil)

					state, err := p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.RequiredAction, ShouldEqual, interaction.StepActionSetupAuthenticator)
					So(state.AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.AvailableAuthenticators[0], ShouldResemble, interaction.AuthenticatorSpec{
						Type:  interaction.AuthenticatorTypePassword,
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
					So(state.RequiredAction, ShouldEqual, interaction.StepActionAuthenticatePrimary)
					So(state.AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.AvailableAuthenticators[0], ShouldResemble, interaction.AuthenticatorSpec{
						Type:  interaction.AuthenticatorTypePassword,
						Props: map[string]interface{}{},
					})

					err = p.PerformAction(i, interaction.ActionAuthenticate{
						Authenticator: state.AvailableAuthenticators[0],
						Secret:        "password",
					})
					So(err, ShouldBeNil)

					state, err = p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.RequiredAction, ShouldEqual, interaction.StepActionCompleted)

					err = p.Commit(i)
					So(err, ShouldBeNil)
				})
			})

			Convey("Login", func() {
				Convey("step 1", func() {
					i, err := p.NewInteraction(
						interaction.IntentLogin{Identity: interaction.IdentitySpec{
							Type:   interaction.IdentityTypeLoginID,
							Claims: map[string]interface{}{"email": "user@example.com"},
						}},
						nil,
					)
					So(err, ShouldBeNil)

					state, err := p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.RequiredAction, ShouldEqual, interaction.StepActionAuthenticatePrimary)
					So(state.AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.AvailableAuthenticators[0], ShouldResemble, interaction.AuthenticatorSpec{
						Type:  interaction.AuthenticatorTypePassword,
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
					So(state.RequiredAction, ShouldEqual, interaction.StepActionAuthenticatePrimary)
					So(state.AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.AvailableAuthenticators[0], ShouldResemble, interaction.AuthenticatorSpec{
						Type:  interaction.AuthenticatorTypePassword,
						Props: map[string]interface{}{},
					})

					err = p.PerformAction(i, interaction.ActionAuthenticate{
						Authenticator: state.AvailableAuthenticators[0],
						Secret:        "password",
					})
					So(err, ShouldBeNil)

					state, err = p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.RequiredAction, ShouldEqual, interaction.StepActionCompleted)

					err = p.Commit(i)
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("SSO flow with MFA", func() {
			Convey("step 1", func() {
				i, err := p.NewInteraction(
					interaction.IntentLogin{
						Identity: interaction.IdentitySpec{
							Type: interaction.IdentityTypeOAuth,
							Claims: map[string]interface{}{
								interaction.IdentityClaimOAuthProvider: map[string]interface{}{
									"type":   "azureadv2",
									"tenant": "example",
								},
								interaction.IdentityClaimOAuthSubjectID: "9A8822AA-4F18-4E4C-84AF-E0FD9AB86CB2",
								interaction.IdentityClaimOAuthProfile:   map[string]interface{}{},
							},
						},
					},
					nil,
				)
				So(err, ShouldBeNil)

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.RequiredAction, ShouldEqual, interaction.StepActionAuthenticateSecondary)
				So(state.AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.AvailableAuthenticators[0], ShouldResemble, interaction.AuthenticatorSpec{
					Type: interaction.AuthenticatorTypeTOTP,
					Props: map[string]interface{}{
						interaction.AuthenticatorPropTOTPDisplayName: "My Authenticator",
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
				So(state.RequiredAction, ShouldEqual, interaction.StepActionAuthenticateSecondary)
				So(state.AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.AvailableAuthenticators[0], ShouldResemble, interaction.AuthenticatorSpec{
					Type: interaction.AuthenticatorTypeTOTP,
					Props: map[string]interface{}{
						interaction.AuthenticatorPropTOTPDisplayName: "My Authenticator",
					},
				})

				err = p.PerformAction(i, interaction.ActionAuthenticate{
					Authenticator: state.AvailableAuthenticators[0],
					Secret:        "123456",
				})
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.RequiredAction, ShouldEqual, interaction.StepActionCompleted)

				err = p.Commit(i)
				So(err, ShouldBeNil)
			})
		})

		Convey("Setup MFA", func() {
			Convey("step 1", func() {
				i, err := p.NewInteraction(
					interaction.IntentAddAuthenticator{
						Authenticator: interaction.AuthenticatorSpec{
							Type: interaction.AuthenticatorTypeTOTP,
							Props: map[string]interface{}{
								interaction.AuthenticatorPropTOTPDisplayName: "My Authenticator",
							},
						},
					},
					nil,
				)
				So(err, ShouldBeNil)

				So(i.PendingAuthenticator, ShouldNotBeEmpty)
				So(i.PendingAuthenticator, ShouldResemble, interaction.AuthenticatorSpec{
					Type: interaction.AuthenticatorTypeTOTP,
					Props: map[string]interface{}{
						interaction.AuthenticatorPropTOTPDisplayName: "My Authenticator",
					},
				})

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.RequiredAction, ShouldEqual, interaction.StepActionSetupAuthenticator)

				err = p.PerformAction(i, interaction.ActionAuthenticate{
					Secret: "123456",
				})
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.RequiredAction, ShouldEqual, interaction.StepActionCompleted)

				err = p.Commit(i)
				So(err, ShouldBeNil)
			})
		})
	})
}
