package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// A Milestone is a marker.
// The current use case that a workflow find if a particular milestone exists,
// to determine whether the workflow has finished.
type Milestone interface {
	Milestone()
}

type MilestoneNestedSteps interface {
	Milestone
	MilestoneNestedSteps()
}

type MilestoneIdentificationMethod interface {
	Milestone
	MilestoneIdentificationMethod() (config.WorkflowIdentificationMethod, bool)
}

type MilestoneAuthenticationMethod interface {
	Milestone
	MilestoneAuthenticationMethod() (config.WorkflowAuthenticationMethod, bool)
}

type MilestoneDoCreateSession interface {
	Milestone
	MilestoneDoCreateSession() bool
}

type MilestoneDoCreateUser interface {
	Milestone
	MilestoneDoCreateUser() (string, bool)
}

type MilestoneDoCreateIdentity interface {
	Milestone
	MilestoneDoCreateIdentity() (*identity.Info, bool)
}

type MilestoneDoCreateAuthenticator interface {
	Milestone
	MilestoneDoCreateAuthenticator() (*authenticator.Info, bool)
}

type MilestoneDoUseIdentity interface {
	Milestone
	MilestoneDoUseIdentity() (*identity.Info, bool)
}

type MilestoneDoUseAuthenticator interface {
	Milestone
	MilestoneDoUseAuthenticator() (*NodeDoUseAuthenticator, bool)
}

type MilestoneDoPopulateStandardAttributes interface {
	Milestone
	MilestoneDoPopulateStandardAttributes()
}
