package authenticationflow

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

// ErrIncompatibleInput means the input reactor cannot react to the input.
// This error can only be returned by ReactTo.
var ErrIncompatibleInput = errors.New("incompatible input")

// ErrSameNode means the input is reacted to, but no node is produced.
// This typically means the node has performed some immediate side effects.
// This error can only be returned by ReactTo.
var ErrSameNode = errors.New("same node")

// ErrReplaceNode means the input is reacted to, but instead of producing a new node to be appended,
// the returned node should replace the latest node.
// This error can only be returned by ReactTo.
var ErrReplaceNode = errors.New("replace node")

// ErrNoChange means the input does not cause the flow to change.
// This error originates from Accept and will be propagated to public API.
var ErrNoChange = errors.New("no change")

// ErrPauseAndRetryAccept means the authflow must pause at this point, process all delayed functions
// And finally, retry the accept process
var ErrPauseAndRetryAccept = errors.New("pause and retry")

// ErrEOF means end of flow.
// This error originates from CanReactTo and will be propagated to public API.
var ErrEOF = errors.New("eof")

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowNotFound").New("flow not found")

var ErrFlowNotAllowed = apierrors.Forbidden.WithReason("AuthenticationFlowNotAllowed").New("flow not allowed")

var ErrStepNotFound = apierrors.NotFound.WithReason("AuthenticationFlowStepNotFound").New("step not found")

var ErrUnknownFlow = apierrors.BadRequest.WithReason("AuthenticationFlowUnknownFlow").New("unknown flow")

var ErrDifferentUserID = apierrors.BadRequest.WithReason("AuthenticationFlowDifferentUserID").New("different user ID")
var ErrNoUserID = apierrors.BadRequest.WithReason("AuthenticationFlowNoUserID").New("no user ID")

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
	// Cause is the error reported by the bot protection provider, if any.
	// It is kept for logging only; it is never shown to the end user.
	// Note that Cause is intentionally NOT exposed via Unwrap,
	// so that this special error keeps being matched by Status only.
	Cause error
}

func (e *ErrorBotProtectionVerification) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("bot protection verification status: %v: %v", e.Status, e.Cause)
	}
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

// NewErrorBotProtectionVerificationFailed is like ErrorBotProtectionVerificationFailed,
// except that it remembers cause.
func NewErrorBotProtectionVerificationFailed(cause error) *ErrorBotProtectionVerification {
	return &ErrorBotProtectionVerification{
		Status: ErrorBotProtectionVerificationStatusFailed,
		Cause:  cause,
	}
}

// NewErrorBotProtectionVerificationServiceUnavailable is like ErrorBotProtectionVerificationServiceUnavailable,
// except that it remembers cause.
func NewErrorBotProtectionVerificationServiceUnavailable(cause error) *ErrorBotProtectionVerification {
	return &ErrorBotProtectionVerification{
		Status: ErrorBotProtectionVerificationStatusServiceUnavailable,
		Cause:  cause,
	}
}

func newAuthenticationFlowError(flows Flows, err error) error {
	return errorutil.WithDetails(err, errorutil.Details{
		"FlowType": apierrors.APIErrorDetail.Value(flows.Nearest.Intent.(PublicFlow).FlowType()),
	})
}
