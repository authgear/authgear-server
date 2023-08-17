package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.WorkflowIdentificationMethod
}

type inputTakeLoginID interface {
	GetLoginID() string
}
