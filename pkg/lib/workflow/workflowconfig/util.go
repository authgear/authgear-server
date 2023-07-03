package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func FindSignupFlow(c *config.WorkflowConfig, id string) (*config.WorkflowSignupFlow, error) {
	for _, f := range c.SignupFlows {
		if f.ID == id {
			f := f
			return f, nil
		}
	}
	return nil, ErrFlowNotFound
}
