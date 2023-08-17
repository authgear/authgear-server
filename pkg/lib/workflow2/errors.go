package workflow2

import (
	"errors"

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

// ErrNoChange means the input does not cause the workflow to change.
// This error originates from Accept and will be propagated to public API.
var ErrNoChange = errors.New("no change")

// ErrEOF means end of workflow.
// This error originates from CanReactTo and will be propagated to public API.
var ErrEOF = errors.New("eof")

var ErrWorkflowNotFound = apierrors.NotFound.WithReason("WorkflowNotFound").New("workflow not found")

var ErrUnknownFlow = apierrors.BadRequest.WithReason("WorkflowUnknownFlow").New("unknown flow")

var ErrUnknownInput = apierrors.BadRequest.WithReason("WorkflowUnknownInput").New("unknown input")

var ErrInvalidInputKind = apierrors.BadRequest.WithReason("WorkflowInvalidInputKind").New("invalid input kind")

var ErrUserAgentUnmatched = apierrors.Forbidden.WithReason("UserAgentUnmatched").New("workflow cannot be used in other user agent")
