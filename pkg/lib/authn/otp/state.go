package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type State struct {
	Target          string
	ExpireAt        time.Time
	CanResendAt     time.Time
	SubmittedCode   string
	UserID          string
	TooManyAttempts bool

	WebSessionID                           string
	WorkflowID                             string
	AuthenticationFlowWebsocketChannelName string
	AuthenticationFlowType                 string
	AuthenticationFlowName                 string
	AuthenticationFlowJSONPointer          jsonpointer.T

	DeliveryStatus model.OTPDeliveryStatus
	DeliveryError  *apierrors.APIError
}
