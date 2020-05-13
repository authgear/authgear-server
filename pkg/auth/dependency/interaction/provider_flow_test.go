package interaction_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/model"
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

			Convey("Login", func() {
				// step 1 setup
				userID := "user_id_1"
				loginIDClaims := map[string]interface{}{"email": "user@example.com"}
				ii := &identity.Info{
					ID:     "identity_id_1",
					Type:   authn.IdentityTypeLoginID,
					Claims: loginIDClaims,
				}
				ai := &authenticator.Info{
					ID:     "authenticator_id_1",
					Type:   authn.AuthenticatorTypePassword,
					Props:  map[string]interface{}{},
					Secret: "password",
				}
				store.EXPECT().Create(gomock.Any()).Return(nil)

				identityProvider.EXPECT().GetByClaims(
					gomock.Eq(authn.IdentityTypeLoginID), gomock.Eq(loginIDClaims),
				).Return(userID, ii, nil).AnyTimes()
				authenticatorProvider.EXPECT().ListByIdentity(
					gomock.Eq(userID), gomock.Eq(ii),
				).Return([]*authenticator.Info{ai}, nil).AnyTimes()

				// step 1
				i, err := p.NewInteractionLogin(
					&interaction.IntentLogin{Identity: identity.Spec{
						Type:   authn.IdentityTypeLoginID,
						Claims: loginIDClaims,
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

				iCopy := *i
				token, err := p.SaveInteraction(i)
				So(err, ShouldBeNil)
				So(token, ShouldNotBeEmpty)

				// step 2 setup
				store.EXPECT().Get(gomock.Eq(token)).Return(&iCopy, nil)
				store.EXPECT().Delete(gomock.Any()).Return(nil)

				authenticatorProvider.EXPECT().Authenticate(
					gomock.Eq(userID), gomock.Eq(ai.ToSpec()), gomock.Any(), gomock.Any(),
				).Return(ai, nil)

				var emptyIdentityInfoList []*identity.Info
				identityProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().Get(gomock.Eq(userID), ii.Type, ii.ID).Return(ii, nil)

				var emptyAuthenticatorInfoList []*authenticator.Info
				authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq(emptyAuthenticatorInfoList)).Return(nil)
				authenticatorProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyAuthenticatorInfoList)).Return(nil)

				// step 2
				i2, err := p.GetInteraction(token)
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i2)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticatePrimary)
				So(state.Steps[0].AvailableAuthenticators, ShouldNotBeEmpty)
				So(state.Steps[0].AvailableAuthenticators[0], ShouldResemble, authenticator.Spec{
					Type:  authn.AuthenticatorTypePassword,
					Props: map[string]interface{}{},
				})

				err = p.PerformAction(i2, interaction.StepAuthenticatePrimary, &interaction.ActionAuthenticate{
					Authenticator: state.Steps[0].AvailableAuthenticators[0],
					Secret:        "password",
				})
				So(err, ShouldBeNil)

				state, err = p.GetInteractionState(i2)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 2)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticatePrimary)
				So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

				_, err = p.Commit(i2)
				So(err, ShouldBeNil)

			})
		})

		Convey("SSO flow with MFA", func() {
			authnConfig := &config.AuthenticationConfiguration{
				SecondaryAuthenticators:     []string{"totp"},
				SecondaryAuthenticationMode: config.SecondaryAuthenticationModeIfExists,
			}
			p.Config = authnConfig

			userID := "user_id_1"
			oauthClaims := map[string]interface{}{
				identity.IdentityClaimOAuthProvider: map[string]interface{}{
					"type":   "azureadv2",
					"tenant": "example",
				},
				identity.IdentityClaimOAuthSubjectID: "9A8822AA-4F18-4E4C-84AF-E0FD9AB86CB2",
				identity.IdentityClaimOAuthProfile:   map[string]interface{}{},
			}
			ii := &identity.Info{
				ID:     "identity_id_1",
				Type:   authn.IdentityTypeOAuth,
				Claims: oauthClaims,
			}
			ai := &authenticator.Info{
				ID:   "authenticator_id_1",
				Type: authn.AuthenticatorTypeTOTP,
				Props: map[string]interface{}{
					"https://auth.skygear.io/claims/totp/display_name": "My Authenticator",
				},
			}

			// step 1 setup
			identityProvider.EXPECT().GetByClaims(
				gomock.Eq(authn.IdentityTypeOAuth), gomock.Eq(oauthClaims),
			).Return(userID, ii, nil).AnyTimes()
			// no primary authenticator for oauth identity
			authenticatorProvider.EXPECT().ListByIdentity(
				gomock.Eq(userID), gomock.Eq(ii),
			).Return([]*authenticator.Info{}, nil).AnyTimes()
			// simulate user has setup totp authenticator
			authenticatorProvider.EXPECT().List(
				gomock.Eq(userID), gomock.Eq(authn.AuthenticatorTypeTOTP),
			).Return([]*authenticator.Info{ai}, nil).AnyTimes()
			store.EXPECT().Create(gomock.Any()).Return(nil)

			// step 1
			i, err := p.NewInteractionLogin(
				&interaction.IntentLogin{
					Identity: identity.Spec{
						Type:   authn.IdentityTypeOAuth,
						Claims: oauthClaims,
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

			iCopy := *i
			token, err := p.SaveInteraction(i)
			So(err, ShouldBeNil)
			So(token, ShouldNotBeEmpty)

			// step 2 setup
			store.EXPECT().Get(gomock.Eq(token)).Return(&iCopy, nil)
			store.EXPECT().Delete(gomock.Any()).Return(nil)

			identityProvider.EXPECT().Get(gomock.Eq(userID), ii.Type, ii.ID).Return(ii, nil)
			identityProvider.EXPECT().WithClaims(
				gomock.Eq(userID), gomock.Eq(ii), gomock.Eq(oauthClaims),
			).Return(ii)

			// update oauth claims when login
			identityProvider.EXPECT().UpdateAll(gomock.Any(), []*identity.Info{ii}).Return(nil)

			var emptyIdentityInfoList []*identity.Info
			identityProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
			identityProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
			identityProvider.EXPECT().Get(gomock.Eq(userID), ii.Type, ii.ID).Return(ii, nil)

			var emptyAuthenticatorInfoList []*authenticator.Info
			authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq(emptyAuthenticatorInfoList)).Return(nil)
			authenticatorProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyAuthenticatorInfoList)).Return(nil)

			// step 2 authenticate secondary authenticator
			i2, err := p.GetInteraction(token)
			So(err, ShouldBeNil)

			authenticatorProvider.EXPECT().Authenticate(
				gomock.Eq(userID), gomock.Eq(ai.ToSpec()), gomock.Any(), gomock.Any(),
			).Return(ai, nil)

			state, err = p.GetInteractionState(i2)
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

			err = p.PerformAction(i2, interaction.StepAuthenticateSecondary, &interaction.ActionAuthenticate{
				Authenticator: state.Steps[0].AvailableAuthenticators[0],
				Secret:        "123456",
			})
			So(err, ShouldBeNil)

			state, err = p.GetInteractionState(i2)
			So(err, ShouldBeNil)
			So(state.Steps, ShouldHaveLength, 2)
			So(state.Steps[0].Step, ShouldEqual, interaction.StepAuthenticateSecondary)
			So(state.Steps[1].Step, ShouldEqual, interaction.StepCommit)

			_, err = p.Commit(i2)
			So(err, ShouldBeNil)
		})

		SkipConvey("Setup MFA", func() {
			// TODO(interaction): setup secondary authenticator
			var p *interaction.Provider

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

		Convey("Add identity", func() {
			userID := "user_id_1"
			p.Config = &config.AuthenticationConfiguration{
				PrimaryAuthenticators: []string{"password"},
			}

			Convey("should not need to setup authenticator", func() {
				// setup
				loginIDClaims := map[string]interface{}{
					identity.IdentityClaimLoginIDKey:   "email",
					identity.IdentityClaimLoginIDValue: "second@example.com",
				}
				ii := &identity.Info{
					ID:     "identity_id_2",
					Type:   authn.IdentityTypeLoginID,
					Claims: loginIDClaims,
				}
				oii := &identity.Info{
					ID:   "identity_id_1",
					Type: authn.IdentityTypeLoginID,
					Claims: map[string]interface{}{
						identity.IdentityClaimLoginIDKey:   "email",
						identity.IdentityClaimLoginIDValue: "user@example.com",
					},
				}
				ai := &authenticator.Info{
					ID:     "authenticator_id_1",
					Type:   authn.AuthenticatorTypePassword,
					Props:  map[string]interface{}{},
					Secret: "password",
				}
				as := &authenticator.Spec{
					Type:  authn.AuthenticatorTypePassword,
					Props: map[string]interface{}{},
				}
				identityProvider.EXPECT().New(
					gomock.Eq(userID), gomock.Eq(authn.IdentityTypeLoginID), gomock.Eq(loginIDClaims),
				).Return(ii)
				// return user's existing identities
				identityProvider.EXPECT().ListByUser(
					gomock.Eq(userID),
				).Return([]*identity.Info{oii}, nil)
				// should include both old and new identities in validation
				identityProvider.EXPECT().Validate(
					gomock.Eq([]*identity.Info{ii, oii}),
				).Return(nil)

				// return new identity related authenticator spec
				identityProvider.EXPECT().RelateIdentityToAuthenticator(
					gomock.Eq(ii.ToSpec()), gomock.Eq(as),
				).Return(as)

				// user has setup authenticator before, no need to setup
				// authenticator
				authenticatorProvider.EXPECT().ListByIdentity(
					gomock.Eq(userID), gomock.Eq(ii),
				).Return([]*authenticator.Info{ai}, nil)

				// start flow
				i, err := p.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
					Identity: identity.Spec{
						Type:   authn.IdentityTypeLoginID,
						Claims: loginIDClaims,
					},
				}, "", userID)

				So(err, ShouldBeNil)

				state, err := p.GetInteractionState(i)
				So(err, ShouldBeNil)
				So(state.Steps, ShouldHaveLength, 1)
				So(state.Steps[0].Step, ShouldEqual, interaction.StepCommit)
			})

			Convey("should setup authenticator", func() {
				// user who had oauth identity without authenticator add login id identity

				// setup
				loginIDClaims := map[string]interface{}{
					identity.IdentityClaimLoginIDKey:   "email",
					identity.IdentityClaimLoginIDValue: "second@example.com",
				}
				ii := &identity.Info{
					ID:     "identity_id_2",
					Type:   authn.IdentityTypeLoginID,
					Claims: loginIDClaims,
				}
				oii := &identity.Info{
					ID:   "identity_id_1",
					Type: authn.IdentityTypeOAuth,
					Claims: map[string]interface{}{
						identity.IdentityClaimOAuthProvider: map[string]interface{}{
							"type":   "azureadv2",
							"tenant": "example",
						},
						identity.IdentityClaimOAuthSubjectID: "9A8822AA-4F18-4E4C-84AF-E0FD9AB86CB2",
						identity.IdentityClaimOAuthProfile:   map[string]interface{}{},
					},
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
				identityProvider.EXPECT().New(
					gomock.Eq(userID), gomock.Eq(ii.Type), gomock.Eq(ii.Claims),
				).Return(ii)
				// return user's existing identities
				identityProvider.EXPECT().ListByUser(
					gomock.Eq(userID),
				).Return([]*identity.Info{oii}, nil)
				// validation should have the new identity only, since the
				// existing identity is in different type
				identityProvider.EXPECT().Validate(
					gomock.Eq([]*identity.Info{ii}),
				).Return(nil)

				// return new identity related authenticator spec
				identityProvider.EXPECT().RelateIdentityToAuthenticator(
					gomock.Eq(ii.ToSpec()), gomock.Eq(as),
				).Return(as).AnyTimes()

				identityProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq([]*identity.Info{ii})).Return(nil)
				var emptyIdentityInfoList []*identity.Info
				identityProvider.EXPECT().UpdateAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyIdentityInfoList)).Return(nil)
				identityProvider.EXPECT().Get(gomock.Any(), ii.Type, ii.ID).Return(ii, nil)

				// no existing authenticator for new login id identity
				authenticatorProvider.EXPECT().ListByIdentity(
					gomock.Eq(userID), gomock.Eq(ii),
				).Return([]*authenticator.Info{}, nil).AnyTimes()

				authenticatorProvider.EXPECT().New(
					gomock.Any(), gomock.Eq(*as), gomock.Eq("password"),
				).Return([]*authenticator.Info{ai}, nil)
				authenticatorProvider.EXPECT().CreateAll(gomock.Any(), gomock.Eq([]*authenticator.Info{ai})).Return(nil)
				var emptyAuthenticatorInfoList []*authenticator.Info
				authenticatorProvider.EXPECT().DeleteAll(gomock.Any(), gomock.Eq(emptyAuthenticatorInfoList)).Return(nil)

				// get user for hook
				userProvider.EXPECT().Get(gomock.Eq(userID)).Return(&model.User{}, nil)

				store.EXPECT().Delete(gomock.Any()).Return(nil)

				// start flow
				i, err := p.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
					Identity: identity.Spec{
						Type:   authn.IdentityTypeLoginID,
						Claims: loginIDClaims,
					},
				}, "", userID)

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

				// setup primary authenticator
				err = p.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
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
	})

}
