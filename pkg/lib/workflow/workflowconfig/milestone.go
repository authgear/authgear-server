package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

// Milestone is a marker.
// The designed use case is to find out whether a particular milestone exists
// in the workflow, or any of its subworkflows.
type Milestone interface {
	Milestone()
}

func FindMilestone[T Milestone](w *workflow.Workflow) (T, bool) {
	var t T
	found := false

	err := w.Traverse(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, _ *workflow.Workflow) error {
			if m, ok := nodeSimple.(T); ok {
				t = m
				found = true
			}
			return nil
		},
		Intent: func(intent workflow.Intent, w *workflow.Workflow) error {
			if m, ok := intent.(T); ok {
				t = m
				found = true
			}
			return nil
		},
	})
	if err != nil {
		return *new(T), false
	}

	if !found {
		return *new(T), false
	}

	return t, true
}

type MilestoneNestedSteps interface {
	Milestone
	MilestoneNestedSteps()
}

type MilestoneIdentificationMethod interface {
	Milestone
	MilestoneIdentificationMethod() config.WorkflowIdentificationMethod
}

type MilestoneAuthenticationMethod interface {
	Milestone
	MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod
}

type MilestoneDidAuthenticate interface {
	Milestone
	MilestoneDidAuthenticate()
}

type MilestoneDoCreateSession interface {
	Milestone
	MilestoneDoCreateSession() (*idpsession.IDPSession, bool)
}

type MilestoneDoCreateUser interface {
	Milestone
	MilestoneDoCreateUser() string
}

type MilestoneDoCreateIdentity interface {
	Milestone
	MilestoneDoCreateIdentity() *identity.Info
}

type MilestoneDoCreateAuthenticator interface {
	Milestone
	MilestoneDoCreateAuthenticator() *authenticator.Info
}

type MilestoneDoUseIdentity interface {
	Milestone
	MilestoneDoUseIdentity() *identity.Info
}

type MilestoneDidSelectAuthenticator interface {
	Milestone
	MilestoneDidSelectAuthenticator() *authenticator.Info
}

type MilestoneDidVerifyAuthenticator interface {
	Milestone
	MilestoneDidVerifyAuthenticator() *NodeDidVerifyAuthenticator
}

type MilestoneDoPopulateStandardAttributes interface {
	Milestone
	MilestoneDoPopulateStandardAttributes()
}

type MilestoneDoMarkClaimVerified interface {
	Milestone
	MilestoneDoMarkClaimVerified()
}

type MilestoneDeviceTokenInspected interface {
	Milestone
	MilestoneDeviceTokenInspected()
}

type MilestoneDoCreateDeviceTokenIfRequested interface {
	Milestone
	MilestoneDoCreateDeviceTokenIfRequested()
}
