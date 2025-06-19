package declarative_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
)

// MockMilestoneConstraintsProvider implements MilestoneConstraintsProvider interface
type MockMilestoneConstraintsProvider struct {
	amr []string
}

var _ declarative.MilestoneConstraintsProvider = &MockMilestoneConstraintsProvider{}

func (m *MockMilestoneConstraintsProvider) Kind() string {
	return "MockMilestoneConstraintsProvider"
}

func (m *MockMilestoneConstraintsProvider) Milestone() {}

func (m *MockMilestoneConstraintsProvider) MilestoneConstraintsProvider() *eventapi.Constraints {
	return &eventapi.Constraints{
		AMR: m.amr,
	}
}

func (m *MockMilestoneConstraintsProvider) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (m *MockMilestoneConstraintsProvider) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, nil
}

func TestAMRUtils(t *testing.T) {
	Convey("RemainingAMRConstraintsInFlow", t, func() {
		Convey("should find remaining AMR constraints when some are fulfilled", func() {
			// Create a flow with AMR constraints and some fulfilled AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneConstraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP, model.AMRPWD},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseAuthenticatorSimple{
							Authenticator: &authenticator.Info{
								ID:   "auth1",
								Kind: model.AuthenticatorKindPrimary,
								Type: model.AuthenticatorTypePassword,
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldResemble, []string{model.AMRMFA, model.AMROTP})

			amr, err := declarative.CollectAMR(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(amr, ShouldResemble, []string{model.AMRPWD, model.AMRXPrimaryPassword})
		})

		Convey("should return empty when all AMR constraints are fulfilled", func() {
			// Create a flow with AMR constraints and all fulfilled AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneConstraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseAuthenticatorSimple{
							Authenticator: &authenticator.Info{
								ID:   "auth1",
								Kind: model.AuthenticatorKindPrimary,
								Type: model.AuthenticatorTypeOOBEmail,
							},
						},
					},
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseAuthenticatorSimple{
							Authenticator: &authenticator.Info{
								ID:   "auth2",
								Kind: model.AuthenticatorKindPrimary,
								Type: model.AuthenticatorTypePassword,
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldBeEmpty)

			amr, err := declarative.CollectAMR(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(amr, ShouldResemble, []string{model.AMRMFA, model.AMROTP, model.AMRPWD, model.AMRXPrimaryOOBOTPEmail, model.AMRXPrimaryPassword})
		})

		Convey("should find AMR constraints in nested flows", func() {
			// Create a flow with nested flows containing AMR constraints and some fulfilled AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &declarative.IntentLoginFlow{},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSubFlow,
						SubFlow: &authenticationflow.Flow{
							Intent: &MockMilestoneConstraintsProvider{
								amr: []string{model.AMRMFA, model.AMROTP},
							},
							Nodes: []authenticationflow.Node{
								{
									Type: authenticationflow.NodeTypeSimple,
									Simple: &declarative.NodeDoUseAuthenticatorSimple{
										Authenticator: &authenticator.Info{
											ID:   "auth1",
											Kind: model.AuthenticatorKindPrimary,
											Type: model.AuthenticatorTypeOOBSMS,
										},
									},
								},
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldResemble, []string{model.AMRMFA})

			amr, err := declarative.CollectAMR(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(amr, ShouldResemble, []string{model.AMROTP, model.AMRSMS, model.AMRXPrimaryOOBOTPSMS})
		})

		Convey("should gather AMRs from multiple MilestoneDidAuthenticate nodes", func() {
			// Create a flow with multiple nodes that have fulfilled different AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneConstraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP, model.AMRPWD},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseAuthenticatorSimple{
							Authenticator: &authenticator.Info{
								ID:   "auth1",
								Kind: model.AuthenticatorKindPrimary,
								Type: model.AuthenticatorTypePassword,
							},
						},
					},
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseAuthenticatorSimple{
							Authenticator: &authenticator.Info{
								ID:   "auth2",
								Kind: model.AuthenticatorKindPrimary,
								Type: model.AuthenticatorTypeOOBEmail,
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldBeEmpty)

			amr, err := declarative.CollectAMR(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(amr, ShouldResemble, []string{model.AMRMFA, model.AMROTP, model.AMRPWD, model.AMRXPrimaryOOBOTPEmail, model.AMRXPrimaryPassword})
		})

		Convey("should fulfill MFA when recovery code and authenticator are used", func() {
			// Create a flow with AMR constraints and recovery code + authenticator
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneConstraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseAuthenticatorSimple{
							Authenticator: &authenticator.Info{
								ID:   "auth1",
								Kind: model.AuthenticatorKindPrimary,
								Type: model.AuthenticatorTypeOOBEmail,
							},
						},
					},
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoConsumeRecoveryCode{
							RecoveryCode: &mfa.RecoveryCode{
								ID: "rc1",
							},
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldBeEmpty)

			amr, err := declarative.CollectAMR(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(amr, ShouldResemble, []string{model.AMRMFA, model.AMROTP, model.AMRXPrimaryOOBOTPEmail, model.AMRXRecoveryCode})
		})
	})

	Convey("CollectAMR", t, func() {
		Convey("should include x_device_token amr when device_token used", func() {
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneConstraintsProvider{
					amr: []string{},
				},
				Nodes: []authenticationflow.Node{
					{
						Type:   authenticationflow.NodeTypeSimple,
						Simple: &declarative.NodeDoUseDeviceToken{},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			amr, err := declarative.CollectAMR(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(amr, ShouldResemble, []string{model.AMRXDeviceToken})

		})
	})
}
