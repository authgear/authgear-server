package webapp

import "github.com/authgear/authgear-server/pkg/lib/interaction/nodes"

type inputTriggerOOB struct {
	AuthenticatorIndex int
	AuthenticatorType  string
}

var _ nodes.InputAuthenticationOOBTrigger = &inputTriggerOOB{}

func (i *inputTriggerOOB) GetOOBAuthenticatorIndex() int   { return i.AuthenticatorIndex }
func (i *inputTriggerOOB) GetOOBAuthenticatorType() string { return i.AuthenticatorType }

type inputSelectTOTP struct{}

var _ nodes.InputCreateAuthenticatorTOTPSetup = &inputSelectTOTP{}

func (i *inputSelectTOTP) SetupTOTP() {}

type inputAuthDeviceToken struct {
	DeviceToken string
}

var _ nodes.InputUseDeviceToken = &inputAuthDeviceToken{}

func (i *inputAuthDeviceToken) GetDeviceToken() string { return i.DeviceToken }
