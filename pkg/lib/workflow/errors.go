package workflow

import (
	"errors"
)

// ErrIncompatibleInput means the input is incompatible with the edge.
// The input should be feeded to another edge instead.
// This error can only be returned by Instantiate.
var ErrIncompatibleInput = errors.New("incompatible input")

// ErrSameNode means the edge consumed the input, but no node is produced.
// This typically the edge has performed some immediate side effects.
// This error can only be returned by Instantiate.
var ErrSameNode = errors.New("same node")

// ErrUpdateNode means the edge consumed the input, but instead of producing a new node to be appended,
// the returned node should replace the node.
// This error can only be returned by Instantiate.
var ErrUpdateNode = errors.New("update node")

// ErrNoChange means the input does not cause the workflow to change.
// This error originates from Accept and will be propagated to public API.
var ErrNoChange = errors.New("no change")

// ErrEOF means end of workflow.
// This error originates from DeriveEdges and will be propagated to public API.
var ErrEOF = errors.New("eof")
