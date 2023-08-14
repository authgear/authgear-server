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
	workflow.Milestone
	MilestoneNestedSteps()
}

type MilestoneIdentificationMethod interface {
	workflow.Milestone
	MilestoneIdentificationMethod() config.WorkflowIdentificationMethod
}

type MilestoneAuthenticationMethod interface {
	workflow.Milestone
	MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod
}

type MilestoneDidAuthenticate interface {
	workflow.Milestone
	MilestoneDidAuthenticate() (amr []string)
}

type MilestoneDoCreateSession interface {
	workflow.Milestone
	MilestoneDoCreateSession() (*idpsession.IDPSession, bool)
}

type MilestoneDoCreateUser interface {
	workflow.Milestone
	MilestoneDoCreateUser() string
}

type MilestoneDoCreateIdentity interface {
	workflow.Milestone
	MilestoneDoCreateIdentity() *identity.Info
}

type MilestoneDoCreateAuthenticator interface {
	workflow.Milestone
	MilestoneDoCreateAuthenticator() *authenticator.Info
}

type MilestoneDoUseUser interface {
	workflow.Milestone
	MilestoneDoUseUser() string
}

type MilestoneDoUseIdentity interface {
	workflow.Milestone
	MilestoneDoUseIdentity() *identity.Info
}

type MilestoneDidSelectAuthenticator interface {
	workflow.Milestone
	MilestoneDidSelectAuthenticator() *authenticator.Info
}

type MilestoneDidVerifyAuthenticator interface {
	workflow.Milestone
	MilestoneDidVerifyAuthenticator() *NodeDidVerifyAuthenticator
}

type MilestoneDoPopulateStandardAttributes interface {
	workflow.Milestone
	MilestoneDoPopulateStandardAttributes()
}

type MilestoneDoMarkClaimVerified interface {
	workflow.Milestone
	MilestoneDoMarkClaimVerified()
}

type MilestoneDeviceTokenInspected interface {
	workflow.Milestone
	MilestoneDeviceTokenInspected()
}

type MilestoneDoCreateDeviceTokenIfRequested interface {
	workflow.Milestone
	MilestoneDoCreateDeviceTokenIfRequested()
}

type MilestoneDidUseAuthenticationLockoutMethod interface {
	workflow.Milestone
	MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool)
}
