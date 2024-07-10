package authenticationflow

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

// ErrIncompatibleInput means the input reactor cannot react to the input.
// This error can only be returned by ReactTo.
var ErrIncompatibleInput = errors.New("incompatible input")

// ErrSameNode means the input is reacted to, but no node is produced.
// This typically means the node has performed some immediate side effects.
// This error can only be returned by ReactTo.
var ErrSameNode = errors.New("same node")

// ErrUpdateNode means the input is reacted to, but instead of producing a new node to be appended,
// the returned node should replace the node.
// This error can only be returned by ReactTo.
var ErrUpdateNode = errors.New("update node")

// ErrNoChange means the input does not cause the flow to change.
// This error originates from Accept and will be propagated to public API.
var ErrNoChange = errors.New("no change")

// ErrEOF means end of flow.
// This error originates from CanReactTo and will be propagated to public API.
var ErrEOF = errors.New("eof")

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowNotFound").New("flow not found")

var ErrFlowNotAllowed = apierrors.Forbidden.WithReason("AuthenticationFlowNotAllowed").New("flow not allowed")

var ErrStepNotFound = apierrors.NotFound.WithReason("AuthenticationFlowStepNotFound").New("step not found")

var ErrUnknownFlow = apierrors.BadRequest.WithReason("AuthenticationFlowUnknownFlow").New("unknown flow")

// ErrorSwitchFlow is a special error for switching flow.
type ErrorSwitchFlow struct {
	// FlowReference indicates the flow to switch to.
	FlowReference FlowReference
	// SyntheticInput advance the switched flow at the current state.
	// It MUST include the input that triggers this error.
	SyntheticInput Input
}

func (e *ErrorSwitchFlow) Error() string {
	return fmt.Sprintf("switch flow: %v %v", e.FlowReference.Type, e.FlowReference.Name)
}

// ErrorRewriteFlow is a special error for rewriting the flow.
type ErrorRewriteFlow struct {
	Intent Intent
	Nodes  []Node
	// SyntheticInput advance the rewritten flow at the current state.
	SyntheticInput Input
}

func (e *ErrorRewriteFlow) Error() string {
	return fmt.Sprintf("rewrite flow: %v", e.Intent.Kind())
}

// ErrorBotProtectionVerification is a special error for interrupting the flow in case of failed or service-unavailable
type ErrorBotProtectionVerification struct {
	Status ErrorBotProtectionVerificationStatus
}

func (e *ErrorBotProtectionVerification) Error() string {
	return fmt.Sprintf("bot protection verification status: %v", e.Status)
}

type ErrorBotProtectionVerificationStatus string

const (
	ErrorBotProtectionVerificationStatusFailed             ErrorBotProtectionVerificationStatus = "failed"
	ErrorBotProtectionVerificationStatusSuccess            ErrorBotProtectionVerificationStatus = "success"
	ErrorBotProtectionVerificationStatusServiceUnavailable ErrorBotProtectionVerificationStatus = "service-unavailable"
)

var (
	ErrorBotProtectionVerificationFailed *ErrorBotProtectionVerification = &ErrorBotProtectionVerification{
		Status: ErrorBotProtectionVerificationStatusFailed,
	}
	ErrorBotProtectionVerificationSuccess *ErrorBotProtectionVerification = &ErrorBotProtectionVerification{
		Status: ErrorBotProtectionVerificationStatusSuccess,
	}
	ErrorBotProtectionVerificationServiceUnavailable *ErrorBotProtectionVerification = &ErrorBotProtectionVerification{
		Status: ErrorBotProtectionVerificationStatusServiceUnavailable,
	}
)
