package authenticationflow

import (
	"context"
)

// Data is a marker.
// Ensure all data is a struct, not an ad-hoc map.
type Data interface {
	Data()
}

// DataOutputer is an InputReactor.
// The data it outputs allow the caller to proceed.
type DataOutputer interface {
	InputReactor
	OutputData(ctx context.Context, deps *Dependencies, flows Flows) (Data, error)
}

// EndOfFlowDataOutputer is an optional interface to be implemented by PublicFlow.
// The implementation MUST return a Data that contains baseData.
type EndOfFlowDataOutputer interface {
	PublicFlow
	OutputEndOfFlowData(ctx context.Context, deps *Dependencies, flows Flows, baseData *DataFinishRedirectURI) (Data, error)
}

type mapData map[string]interface{}

var _ Data = mapData{}

func (m mapData) Data() {}

type DataFinishRedirectURI struct {
	FinishRedirectURI string `json:"finish_redirect_uri,omitempty"`
}

var _ Data = &DataFinishRedirectURI{}

func (*DataFinishRedirectURI) Data() {}
