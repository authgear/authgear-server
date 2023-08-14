package workflowconfig

import (
	"context"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/slice"
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

	err := workflow.TraverseWorkflow(workflow.WorkflowTraverser{
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
	}, w)
	if err != nil {
		return *new(T), false
	}

	if !found {
		return *new(T), false
	}

	return t, true
}

func getUserID(workflows workflow.Workflows) (userID string, err error) {
	err = workflow.TraverseWorkflow(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, w *workflow.Workflow) error {
			if n, ok := nodeSimple.(MilestoneDoUseUser); ok {
				id := n.MilestoneDoUseUser()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
		Intent: func(intent workflow.Intent, w *workflow.Workflow) error {
			if i, ok := intent.(MilestoneDoUseUser); ok {
				id := i.MilestoneDoUseUser()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
	}, workflows.Root)

	if userID == "" {
		err = ErrNoUserID
	}

	if err != nil {
		return
	}

	return
}

func collectAMR(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (amr []string, err error) {
	err = workflow.TraverseWorkflow(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, w *workflow.Workflow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				amr = append(amr, n.MilestoneDidAuthenticate()...)
			}
			return nil
		},
		Intent: func(intent workflow.Intent, w *workflow.Workflow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				amr = append(amr, i.MilestoneDidAuthenticate()...)
			}
			return nil
		},
	}, workflows.Root)
	if err != nil {
		return
	}

	amr = slice.Deduplicate(amr)
	sort.Strings(amr)

	return
}

func collectAuthenticationLockoutMethod(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (methods []config.AuthenticationLockoutMethod, err error) {
	err = workflow.TraverseWorkflow(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, w *workflow.Workflow) error {
			if n, ok := nodeSimple.(MilestoneDidUseAuthenticationLockoutMethod); ok {
				if m, ok := n.MilestoneDidUseAuthenticationLockoutMethod(); ok {
					methods = append(methods, m)
				}

			}
			return nil
		},
		Intent: func(intent workflow.Intent, w *workflow.Workflow) error {
			if i, ok := intent.(MilestoneDidUseAuthenticationLockoutMethod); ok {
				if m, ok := i.MilestoneDidUseAuthenticationLockoutMethod(); ok {
					methods = append(methods, m)
				}
			}
			return nil
		},
	}, workflows.Root)
	if err != nil {
		return
	}

	return
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
	MilestoneDidAuthenticate() (amr []string)
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

type MilestoneDoUseUser interface {
	Milestone
	MilestoneDoUseUser() string
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

type MilestoneDidUseAuthenticationLockoutMethod interface {
	Milestone
	MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool)
}
