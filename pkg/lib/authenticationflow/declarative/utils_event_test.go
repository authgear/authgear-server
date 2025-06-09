package declarative_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func TestGetAuthenticationContext(t *testing.T) {
	Convey("GetAuthenticationContext", t, func() {
		ctx := context.Background()
		Convey("in login flow", func() {

			fixedTime := time.Date(2025, time.June, 4, 6, 45, 37, 0, time.UTC)

			userID := "test-user-1"
			testUser := &model.User{
				Meta: model.Meta{ID: userID},
			}

			assertedIdentity := &identity.Info{
				ID:        "test-identity-1",
				UserID:    userID,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.IdentityTypeLoginID,
				LoginID: &identity.LoginID{
					LoginIDKey:      "email",
					LoginIDType:     model.LoginIDKeyTypeEmail,
					LoginID:         "test@example.com",
					OriginalLoginID: "test@example.com",
					Claims: map[string]interface{}{
						string(model.ClaimEmail): "test@example.com",
					},
				},
			}

			assertedAuthenticator := &authenticator.Info{
				ID:        "test-authn-1",
				UserID:    userID,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.AuthenticatorTypeOOBEmail,
				Kind:      authenticator.KindPrimary,
				OOBOTP: &authenticator.OOBOTP{
					Email: "test@example.com",
				},
			}

			// Add a second authenticator for MFA test
			assertedAuthenticator2 := &authenticator.Info{
				ID:        "test-authn-2",
				UserID:    userID,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.AuthenticatorTypePassword,
				Kind:      authenticator.KindPrimary,
				Password:  &authenticator.Password{},
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserService := authenticationflow.NewMockUserService(ctrl)
			mockUserService.
				EXPECT().
				Get(gomock.Any(), gomock.Eq(userID), gomock.Any()).
				Return(testUser, nil).
				AnyTimes()

			// Create a mock dependencies instance
			mockDeps := &authenticationflow.Dependencies{
				Users: mockUserService,
			}

			// Construct Flow directly with the nested structure inlined
			rootFlow := &authenticationflow.Flow{
				Intent: &declarative.IntentLoginFlow{
					FlowReference: authenticationflow.FlowReference{
						Type: authenticationflow.FlowTypeLogin,
						Name: "test",
					},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSubFlow,
						SubFlow: &authenticationflow.Flow{
							Intent: &declarative.IntentLoginFlowSteps{
								FlowReference: authenticationflow.FlowReference{},
							},
							Nodes: []authenticationflow.Node{
								{
									Type: authenticationflow.NodeTypeSubFlow,
									SubFlow: &authenticationflow.Flow{
										Intent: &declarative.IntentLoginFlowStepIdentify{
											FlowReference: authenticationflow.FlowReference{},
											StepName:      "stepidentify",
										},
										Nodes: []authenticationflow.Node{
											{
												Type: authenticationflow.NodeTypeSubFlow,
												SubFlow: &authenticationflow.Flow{
													Intent: &declarative.IntentUseIdentityLoginID{},
													Nodes: []authenticationflow.Node{
														{
															Type: authenticationflow.NodeTypeSimple,
															Simple: &declarative.NodeDoUseIdentity{
																Identity: assertedIdentity,
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Type: authenticationflow.NodeTypeSubFlow,
									SubFlow: &authenticationflow.Flow{
										Intent: &declarative.IntentLoginFlowStepAuthenticate{
											FlowReference: authenticationflow.FlowReference{},
											StepName:      "stepauthenticate",
										},
										Nodes: []authenticationflow.Node{
											{
												Type: authenticationflow.NodeTypeSubFlow,
												SubFlow: &authenticationflow.Flow{
													Intent: &declarative.IntentUseAuthenticatorOOBOTP{},
													Nodes: []authenticationflow.Node{
														{
															Type: authenticationflow.NodeTypeSimple,
															Simple: &declarative.NodeDidSelectAuthenticator{
																Authenticator: assertedAuthenticator,
															},
														},
														{
															Type: authenticationflow.NodeTypeSimple,
															Simple: &declarative.NodeDoUseAuthenticatorSimple{
																Authenticator: assertedAuthenticator,
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Type: authenticationflow.NodeTypeSubFlow,
									SubFlow: &authenticationflow.Flow{
										Intent: &declarative.IntentLoginFlowStepAuthenticate{
											FlowReference: authenticationflow.FlowReference{},
											StepName:      "stepauthenticate2",
										},
										Nodes: []authenticationflow.Node{
											{
												Type: authenticationflow.NodeTypeSubFlow,
												SubFlow: &authenticationflow.Flow{
													Intent: &declarative.IntentUseAuthenticatorPassword{},
													Nodes: []authenticationflow.Node{
														{
															Type: authenticationflow.NodeTypeSimple,
															Simple: &declarative.NodeDidSelectAuthenticator{
																Authenticator: assertedAuthenticator2,
															},
														},
														{
															Type: authenticationflow.NodeTypeSimple,
															Simple: &declarative.NodeDoUseAuthenticatorSimple{
																Authenticator: assertedAuthenticator2,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)

			result, err := declarative.GetAuthenticationContext(ctx, mockDeps, flows)
			So(err, ShouldBeNil)
			So(result.AuthenticationFlow, ShouldNotBeNil)
			So(result.AuthenticationFlow.Type, ShouldEqual, string(authenticationflow.FlowTypeLogin))
			So(result.AuthenticationFlow.Name, ShouldEqual, "test")
			So(result.User, ShouldResemble, testUser)
			So(result.AMR, ShouldResemble, []string{"mfa", "otp", "pwd"})
			So(result.AssertedAuthenticators, ShouldHaveLength, 2)
			So(result.AssertedAuthenticators, ShouldContain, assertedAuthenticator.ToModel())
			So(result.AssertedAuthenticators, ShouldContain, assertedAuthenticator2.ToModel())
			So(result.AssertedIdentities, ShouldHaveLength, 1)
			So(result.AssertedIdentities[0], ShouldResemble, assertedIdentity.ToModel())

		})

		Convey("in signup flow", func() {
			fixedTime := time.Date(2025, time.June, 4, 6, 45, 37, 0, time.UTC)

			userID := "test-user-1"
			user := &model.User{
				Meta: model.Meta{ID: userID},
			}

			assertedIdentity := &identity.Info{
				ID:        "test-identity-1",
				UserID:    userID,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.IdentityTypeLoginID,
				LoginID: &identity.LoginID{
					LoginIDKey:      "email",
					LoginIDType:     model.LoginIDKeyTypeEmail,
					LoginID:         "test@example.com",
					OriginalLoginID: "test@example.com",
					Claims: map[string]interface{}{
						string(model.ClaimEmail): "test@example.com",
					},
				},
			}

			assertedAuthenticator := &authenticator.Info{
				ID:        "test-authn-1",
				UserID:    userID,
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.AuthenticatorTypePassword,
				Kind:      authenticator.KindPrimary,
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserService := authenticationflow.NewMockUserService(ctrl)
			mockUserService.
				EXPECT().
				Get(gomock.Any(), gomock.Eq(userID), gomock.Any()).
				Return(user, nil).
				AnyTimes()

			// Create a mock dependencies instance
			mockDeps := &authenticationflow.Dependencies{
				Users: mockUserService,
			}

			// Construct Flow directly with the nested structure inlined
			rootFlow := &authenticationflow.Flow{
				Intent: &declarative.IntentSignupFlow{
					FlowReference: authenticationflow.FlowReference{
						Type: authenticationflow.FlowTypeSignup,
						Name: "test_signup",
					},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoCreateUser{
							UserID: userID,
						},
					},
					{
						Type: authenticationflow.NodeTypeSubFlow,
						SubFlow: &authenticationflow.Flow{
							Intent: &declarative.IntentSignupFlowSteps{
								FlowReference: authenticationflow.FlowReference{},
							},
							Nodes: []authenticationflow.Node{
								{
									Type: authenticationflow.NodeTypeSubFlow,
									SubFlow: &authenticationflow.Flow{
										Intent: &declarative.IntentSignupFlowStepIdentify{
											FlowReference: authenticationflow.FlowReference{},
											StepName:      "stepidentify",
										},
										Nodes: []authenticationflow.Node{
											{
												Type: authenticationflow.NodeTypeSubFlow,
												SubFlow: &authenticationflow.Flow{
													Intent: &declarative.IntentCreateIdentityLoginID{},
													Nodes: []authenticationflow.Node{
														{
															Type: authenticationflow.NodeTypeSubFlow,
															SubFlow: &authenticationflow.Flow{
																Intent: &declarative.IntentCheckConflictAndCreateIdenity{},
																Nodes: []authenticationflow.Node{
																	{
																		Type: authenticationflow.NodeTypeSimple,
																		Simple: &declarative.NodeDoCreateIdentity{
																			Identity: assertedIdentity,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
								{
									Type: authenticationflow.NodeTypeSubFlow,
									SubFlow: &authenticationflow.Flow{
										Intent: &declarative.IntentSignupFlowStepCreateAuthenticator{
											FlowReference: authenticationflow.FlowReference{},
											StepName:      "stepcreateauthenticator",
										},
										Nodes: []authenticationflow.Node{
											{
												Type: authenticationflow.NodeTypeSubFlow,
												SubFlow: &authenticationflow.Flow{
													Intent: &declarative.IntentCreateAuthenticatorPassword{},
													Nodes: []authenticationflow.Node{
														{
															Type: authenticationflow.NodeTypeSimple,
															Simple: &declarative.NodeDoCreateAuthenticator{
																Authenticator: assertedAuthenticator,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)

			result, err := declarative.GetAuthenticationContext(ctx, mockDeps, flows)
			So(err, ShouldBeNil)
			So(result.AuthenticationFlow, ShouldNotBeNil)
			So(result.AuthenticationFlow.Type, ShouldEqual, string(authenticationflow.FlowTypeSignup))
			So(result.AuthenticationFlow.Name, ShouldEqual, "test_signup")
			So(result.User, ShouldResemble, user)
			So(result.AMR, ShouldResemble, []string{"pwd"})
			So(result.AssertedAuthenticators, ShouldHaveLength, 1)
			So(result.AssertedAuthenticators[0], ShouldResemble, assertedAuthenticator.ToModel())
			So(result.AssertedIdentities, ShouldHaveLength, 1)
			So(result.AssertedIdentities[0], ShouldResemble, assertedIdentity.ToModel())

		})
	})
}
