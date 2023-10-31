package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type StepAuthenticateData struct {
	Options            []AuthenticateOptionForOutput `json:"options"`
	DeviceTokenEnabled bool                          `json:"device_token_enable"`
}

var _ authflow.Data = StepAuthenticateData{}

func (m StepAuthenticateData) Data() {}
