package otp

import (
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

type State struct {
	Target                string
	CanResendAt           time.Time
	CanCheckSubmittedCode bool
	UserID                string
	TooManyAttempts       bool

	WebSessionID                           string
	WorkflowID                             string
	AuthenticationFlowWebsocketChannelName string
	AuthenticationFlowType                 string
	AuthenticationFlowName                 string
	AuthenticationFlowJSONPointer          jsonpointer.T

	DeliveryStatus model.OTPDeliveryStatus
	DeliveryError  *apierrors.APIError
}
