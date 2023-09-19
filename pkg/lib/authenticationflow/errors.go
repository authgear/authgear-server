package authenticationflow

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

// ErrNoChange means the input does not cause the flow to change.
// This error originates from Accept and will be propagated to public API.
var ErrNoChange = errors.New("no change")

// ErrEOF means end of flow.
// This error originates from CanReactTo and will be propagated to public API.
var ErrEOF = errors.New("eof")

var ErrFlowNotFound = apierrors.NotFound.WithReason("AuthenticationFlowNotFound").New("flow not found")

var ErrStepNotFound = apierrors.NotFound.WithReason("AuthenticationFlowStepNotFound").New("step not found")

var ErrUnknownFlow = apierrors.BadRequest.WithReason("AuthenticationFlowUnknownFlow").New("unknown flow")
