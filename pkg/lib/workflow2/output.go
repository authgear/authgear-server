package workflow2

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
	OutputData(ctx context.Context, deps *Dependencies, workflows Workflows) (Data, error)
}

type mapData map[string]interface{}

var _ Data = mapData{}

func (m mapData) Data() {}

var EmptyData = make(mapData)

type DataRedirectURI struct {
	RedirectURI string `json:"redirect_uri,omitempty"`
}

var _ Data = &DataRedirectURI{}

func (*DataRedirectURI) Data() {}
