package declarative_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	. "github.com/smartystreets/goconvey/convey"
)

// MockUserService is a mock implementation of authenticationflow.UserService
type MockUserService struct {
	MockGet func(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

func (m *MockUserService) Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
	if m.MockGet != nil {
		return m.MockGet(ctx, id, role)
	}
	return nil, errors.New("Get not implemented")
}

func (m *MockUserService) GetRaw(ctx context.Context, id string) (*user.User, error) {
	return nil, errors.New("GetRaw not implemented")
}

func (m *MockUserService) Create(ctx context.Context, userID string) (*user.User, error) {
	return nil, errors.New("Create not implemented")
}

func (m *MockUserService) UpdateLoginTime(ctx context.Context, userID string, t time.Time) error {
	return errors.New("UpdateLoginTime not implemented")
}

func (m *MockUserService) UpdateMFAEnrollment(ctx context.Context, userID string, t *time.Time) error {
	return errors.New("UpdateMFAEnrollment not implemented")
}

func (m *MockUserService) AfterCreate(
	ctx context.Context,
	user *user.User,
	identities []*identity.Info,
	authenticators []*authenticator.Info,
	isAdminAPI bool,
) error {
	return errors.New("AfterCreate not implemented")
}

func TestGetAuthenticationContext(t *testing.T) {
	Convey("GetAuthenticationContext", t, func() {
		ctx := context.Background()
		Convey("in login flow", func() {

			fixedTime := time.Date(2025, time.June, 4, 6, 45, 37, 0, time.UTC)

			testUser := &model.User{
				Meta: model.Meta{
					ID: "testuserid",
				},
			}

			assertedIdentity := &identity.Info{
				ID:        "test-identity-1",
				UserID:    "test-user-1",
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
				UserID:    "test-user-1",
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.AuthenticatorTypeOOBEmail,
				Kind:      authenticator.KindPrimary,
				OOBOTP: &authenticator.OOBOTP{
					Email: "test@example.com",
				},
			}

			// Create a mock dependencies instance
			mockDeps := &authenticationflow.Dependencies{
				Users: &MockUserService{
					MockGet: func(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
						return testUser, nil
					},
				},
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
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)

			result, err := declarative.GetAuthenticationContext(ctx, flows, mockDeps)
			So(err, ShouldBeNil)
			So(result.AuthenticationFlow, ShouldNotBeNil)
			So(result.AuthenticationFlow.Type, ShouldEqual, string(authenticationflow.FlowTypeLogin))
			So(result.AuthenticationFlow.Name, ShouldEqual, "test")
			So(result.User, ShouldResemble, testUser)
			So(result.AMR, ShouldResemble, []string{"otp"})
			So(result.AssertedAuthenticators, ShouldHaveLength, 1)
			So(result.AssertedAuthenticators[0], ShouldResemble, assertedAuthenticator.ToModel())
			So(result.AssertedIdentities, ShouldHaveLength, 1)
			So(result.AssertedIdentities[0], ShouldResemble, assertedIdentity.ToModel())

		})

		Convey("in signup flow", func() {
			fixedTime := time.Date(2025, time.June, 4, 6, 45, 37, 0, time.UTC)

			assertedIdentity := &identity.Info{
				ID:        "test-identity-1",
				UserID:    "test-user-1", // UserID might be empty or a temporary ID during signup
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
				UserID:    "test-user-1", // UserID might be empty or a temporary ID during signup
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
				Type:      model.AuthenticatorTypePassword,
				Kind:      authenticator.KindPrimary,
				// Password authenticator doesn't have specific data in Info struct
			}

			// Create a mock dependencies instance
			mockDeps := &authenticationflow.Dependencies{
				Users: &MockUserService{
					MockGet: func(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
						// For signup flow, we expect the user not to exist initially
						return nil, user.ErrUserNotFound
					},
				},
			}

			// Construct Flow directly with the nested structure inlined
			rootFlow := &authenticationflow.Flow{
				Intent: &declarative.IntentSignupFlow{
					FlowReference: authenticationflow.FlowReference{
						Type: authenticationflow.FlowTypeSignup,
						Name: "test_signup", // Use a different name for clarity
					},
				},
				Nodes: []authenticationflow.Node{
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

			result, err := declarative.GetAuthenticationContext(ctx, flows, mockDeps)
			So(err, ShouldBeNil)
			So(result.AuthenticationFlow, ShouldNotBeNil)
			So(result.AuthenticationFlow.Type, ShouldEqual, string(authenticationflow.FlowTypeSignup))
			So(result.AuthenticationFlow.Name, ShouldEqual, "test_signup")
			So(result.User, ShouldBeNil) // User should be nil in signup flow initially
			So(result.AMR, ShouldBeEmpty)
			So(result.AssertedAuthenticators, ShouldHaveLength, 1)
			So(result.AssertedAuthenticators[0], ShouldResemble, assertedAuthenticator.ToModel())
			So(result.AssertedIdentities, ShouldHaveLength, 1)
			So(result.AssertedIdentities[0], ShouldResemble, assertedIdentity.ToModel())

		})
	})
}
