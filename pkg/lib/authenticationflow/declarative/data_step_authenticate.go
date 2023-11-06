package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type stepAuthenticateData struct {
	Options            []AuthenticateOptionForOutput `json:"options"`
	DeviceTokenEnabled bool                          `json:"device_token_enable"`
}

var _ authflow.Data = stepAuthenticateData{}

func (m stepAuthenticateData) Data() {}
