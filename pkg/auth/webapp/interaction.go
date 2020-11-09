package webapp

import "github.com/authgear/authgear-server/pkg/lib/interaction/nodes"

type inputTriggerOOB struct{ AuthenticatorIndex int }

var _ nodes.InputAuthenticationOOBTrigger = &inputTriggerOOB{}

func (i *inputTriggerOOB) GetOOBAuthenticatorIndex() int { return i.AuthenticatorIndex }

type inputSelectTOTP struct{}

var _ nodes.InputCreateAuthenticatorTOTPSetup = &inputSelectTOTP{}

func (i *inputSelectTOTP) SetupTOTP() {}

type inputAuthDeviceToken struct {
	DeviceToken string
}

var _ nodes.InputUseDeviceToken = &inputAuthDeviceToken{}

func (i *inputAuthDeviceToken) GetDeviceToken() string { return i.DeviceToken }
