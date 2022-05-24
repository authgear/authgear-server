package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type inputTriggerOOB struct {
	AuthenticatorIndex int
	AuthenticatorType  string
}

var _ nodes.InputAuthenticationOOBTrigger = &inputTriggerOOB{}

func (i *inputTriggerOOB) GetOOBAuthenticatorIndex() int   { return i.AuthenticatorIndex }
func (i *inputTriggerOOB) GetOOBAuthenticatorType() string { return i.AuthenticatorType }

type inputTriggerWhatsapp struct {
	AuthenticatorIndex int
}

var _ nodes.InputAuthenticationWhatsappTrigger = &inputTriggerWhatsapp{}

func (i *inputTriggerWhatsapp) GetWhatsappAuthenticatorIndex() int { return i.AuthenticatorIndex }

type inputSelectTOTP struct{}

var _ nodes.InputCreateAuthenticatorTOTPSetup = &inputSelectTOTP{}

func (i *inputSelectTOTP) SetupTOTP() {}

type inputAuthDeviceToken struct {
	DeviceToken string
}

var _ nodes.InputUseDeviceToken = &inputAuthDeviceToken{}

func (i *inputAuthDeviceToken) GetDeviceToken() string { return i.DeviceToken }

type inputSelectWhatsappOTP struct{}

func (i *inputSelectWhatsappOTP) SetupPrimaryAuthenticatorWhatsappOTP() {}

var _ nodes.InputCreateAuthenticatorWhatsappOTPSetupSelect = &inputSelectWhatsappOTP{}

type inputSelectOOB struct{}

func (i *inputSelectOOB) SetupPrimaryAuthenticatorOOB() {}

var _ nodes.InputCreateAuthenticatorOOBSetupSelect = &inputSelectOOB{}

type inputSelectVerifyIdentityViaOOBOTP struct{}

func (i *inputSelectVerifyIdentityViaOOBOTP) SelectVerifyIdentityViaOOBOTP() {}

var _ nodes.InputVerifyIdentity = &inputSelectVerifyIdentityViaOOBOTP{}

type inputSelectVerifyIdentityViaWhatsapp struct{}

func (i *inputSelectVerifyIdentityViaWhatsapp) SelectVerifyIdentityViaWhatsapp() {}

var _ nodes.InputVerifyIdentityViaWhatsapp = &inputSelectVerifyIdentityViaWhatsapp{}
