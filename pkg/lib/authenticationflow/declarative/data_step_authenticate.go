package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type StepAuthenticateData struct {
	TypedData
	Options            []AuthenticateOptionForOutput `json:"options"`
	DeviceTokenEnabled bool                          `json:"device_token_enabled"`
}

var _ authflow.Data = StepAuthenticateData{}

func (m StepAuthenticateData) Data() {}

func NewStepAuthenticateData(d StepAuthenticateData) StepAuthenticateData {
	d.Type = DataTypeAuthenticationData
	return d
}
