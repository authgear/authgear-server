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

// MockMilestoneContraintsProvider implements MilestoneContraintsProvider interface
type MockMilestoneContraintsProvider struct {
	amr []string
}

func (m *MockMilestoneContraintsProvider) Kind() string {
	return "MockMilestoneContraintsProvider"
}

func (m *MockMilestoneContraintsProvider) Milestone() {}

func (m *MockMilestoneContraintsProvider) MilestoneContraintsProvider() *eventapi.Constraints {
	return &eventapi.Constraints{
		AMR: m.amr,
	}
}

func (m *MockMilestoneContraintsProvider) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (m *MockMilestoneContraintsProvider) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, nil
}

// MockMilestoneDidAuthenticate implements MilestoneDidAuthenticate interface
type MockMilestoneDidAuthenticate struct {
	amr             []string
	authenticatorID string
}

func (m *MockMilestoneDidAuthenticate) Kind() string {
	return "MockMilestoneDidAuthenticate"
}

func (m *MockMilestoneDidAuthenticate) Milestone() {}

func (m *MockMilestoneDidAuthenticate) MilestoneDidAuthenticate() []string {
	return m.amr
}

func (m *MockMilestoneDidAuthenticate) MilestoneDidAuthenticateAuthenticator() (*authenticator.Info, bool) {
	if m.authenticatorID == "" {
		return nil, false
	}
	return &authenticator.Info{
		ID: m.authenticatorID,
	}, true
}

func (m *MockMilestoneDidAuthenticate) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (m *MockMilestoneDidAuthenticate) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, nil
}

// MockMilestoneDidConsumeRecoveryCode implements MilestoneDidConsumeRecoveryCode interface
type MockMilestoneDidConsumeRecoveryCode struct {
	recoveryCodeID string
}

func (m *MockMilestoneDidConsumeRecoveryCode) Kind() string {
	return "MockMilestoneDidConsumeRecoveryCode"
}

func (m *MockMilestoneDidConsumeRecoveryCode) Milestone() {}

func (m *MockMilestoneDidConsumeRecoveryCode) MilestoneDidConsumeRecoveryCode() *mfa.RecoveryCode {
	if m.recoveryCodeID == "" {
		return nil
	}
	return &mfa.RecoveryCode{
		ID: m.recoveryCodeID,
	}
}

func (m *MockMilestoneDidConsumeRecoveryCode) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (m *MockMilestoneDidConsumeRecoveryCode) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, nil
}

func TestRemainingAMRConstraintsInFlow(t *testing.T) {
	Convey("remainingAMRConstraintsInFlow", t, func() {
		Convey("should find remaining AMR constraints when some are fulfilled", func() {
			// Create a flow with AMR constraints and some fulfilled AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneContraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP, model.AMRPWD},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidAuthenticate{
							amr:             []string{model.AMRPWD},
							authenticatorID: "auth1",
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldResemble, []string{model.AMRMFA, model.AMROTP})
		})

		Convey("should return empty when all AMR constraints are fulfilled", func() {
			// Create a flow with AMR constraints and all fulfilled AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneContraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidAuthenticate{
							amr:             []string{model.AMROTP},
							authenticatorID: "auth1",
						},
					},
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidAuthenticate{
							amr:             []string{model.AMRPWD},
							authenticatorID: "auth2",
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldBeEmpty)
		})

		Convey("should find AMR constraints in nested flows", func() {
			// Create a flow with nested flows containing AMR constraints and some fulfilled AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &declarative.IntentLoginFlow{},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSubFlow,
						SubFlow: &authenticationflow.Flow{
							Intent: &MockMilestoneContraintsProvider{
								amr: []string{model.AMRMFA, model.AMROTP},
							},
							Nodes: []authenticationflow.Node{
								{
									Type: authenticationflow.NodeTypeSimple,
									Simple: &MockMilestoneDidAuthenticate{
										amr:             []string{model.AMROTP},
										authenticatorID: "auth1",
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
		})

		Convey("should gather AMRs from multiple MilestoneDidAuthenticate nodes", func() {
			// Create a flow with multiple nodes that have fulfilled different AMRs
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneContraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP, model.AMRPWD},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidAuthenticate{
							amr:             []string{model.AMRPWD},
							authenticatorID: "auth1",
						},
					},
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidAuthenticate{
							amr:             []string{model.AMROTP},
							authenticatorID: "auth2",
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldBeEmpty)
		})

		Convey("should fulfill MFA when recovery code and authenticator are used", func() {
			// Create a flow with AMR constraints and recovery code + authenticator
			rootFlow := &authenticationflow.Flow{
				Intent: &MockMilestoneContraintsProvider{
					amr: []string{model.AMRMFA, model.AMROTP},
				},
				Nodes: []authenticationflow.Node{
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidAuthenticate{
							amr:             []string{model.AMROTP},
							authenticatorID: "auth1",
						},
					},
					{
						Type: authenticationflow.NodeTypeSimple,
						Simple: &MockMilestoneDidConsumeRecoveryCode{
							recoveryCodeID: "rc1",
						},
					},
				},
			}

			flows := authenticationflow.NewFlows(rootFlow)
			constraints, err := declarative.RemainingAMRConstraintsInFlow(context.Background(), nil, flows)
			So(err, ShouldBeNil)
			So(constraints, ShouldBeEmpty)
		})
	})
}
