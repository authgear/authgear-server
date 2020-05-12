package interaction_test

import (
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

func TestProvider(t *testing.T) {
	SkipConvey("Interaction Provider", t, func() {
		var p *interaction.Provider

		Convey("Common password flow", func() {
			Convey("Signup", func() {
				Convey("step 1", func() {
					i, err := p.NewInteractionSignup(
						&interaction.IntentSignup{
							Identity: identity.Spec{
								Type:   authn.IdentityTypeLoginID,
								Claims: map[string]interface{}{"email": "user@example.com"},
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
					So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupPrimaryAuthenticator)
					So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
					So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
						Type:  authn.AuthenticatorTypePassword,
						Props: map[string]interface{}{},
					})

					err = p.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionAuthenticate{
						Authenticator: state.Steps[0].AvailableAuthenticators[0],
						Secret:        "password",
					})
					So(err, ShouldBeNil)

					state, err = p.GetInteractionState(i)
					So(err, ShouldBeNil)
					So(state.Steps, ShouldHaveLength, 2)
					So(state.Steps[0].Step, ShouldEqual, interaction.StepSetupPrimaryAuthenticator)
					So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

					_, err = p.Commit(i)
					So(err, ShouldBeNil)
				})
			})

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

func TestProviderCommit(t *testing.T) {
	Convey("InteractionProviderCommitRemoveIdentity", t, func() {
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
			Store:         store,
			Hooks:         hooks,
		}
		userID := "userid1"
		loginID1 := &identity.Info{
			ID:   "iid1",
			Type: authn.IdentityTypeLoginID,
		}
		loginID2 := &identity.Info{
			ID:   "iid2",
			Type: authn.IdentityTypeLoginID,
		}
		oauthID := &identity.Info{
			ID:   "iid3",
			Type: authn.IdentityTypeOAuth,
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

		store.EXPECT().Create(gomock.Any()).Return(nil).AnyTimes()
		store.EXPECT().Delete(gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().CreateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		identityProvider.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity.Info{}, nil).AnyTimes()
		authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
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

			_, err := p.Commit(i)
			So(err, ShouldBeNil)

			expected := i.RemoveAuthenticators
			actual := []*authenticator.Info{
				pwAuthenticator, totpAuthenticator,
			}
			sort.Sort(authenticatorInfoSlice(expected))
			sort.Sort(authenticatorInfoSlice(actual))
			So(expected, ShouldResemble, actual)

			So(hooks.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{ID: userID},
					Identity: model.Identity{
						Type:   string(loginID1.Type),
						Claims: loginID1.Claims,
					},
				},
			})
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

			_, err := p.Commit(i)
			So(err, ShouldBeNil)

			So(len(i.RemoveAuthenticators), ShouldEqual, 0)

			So(hooks.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{ID: userID},
					Identity: model.Identity{
						Type:   string(oauthID.Type),
						Claims: oauthID.Claims,
					},
				},
			})
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

			So(hooks.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{ID: userID},
					Identity: model.Identity{
						Type:   string(loginID2.Type),
						Claims: loginID2.Claims,
					},
				},
			})
		})
	})
}

type authenticatorInfoSlice []*authenticator.Info

func (s authenticatorInfoSlice) Len() int           { return len(s) }
func (s authenticatorInfoSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s authenticatorInfoSlice) Less(i, j int) bool { return s[i].ID < s[j].ID }
