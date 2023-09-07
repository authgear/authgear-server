package authenticationflow

import (
	"context"
	"encoding/json"
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

type mapData map[string]interface{}

var _ Data = mapData{}

func (m mapData) Data() {}

type DataFinishRedirectURI struct {
	FinishRedirectURI string `json:"finish_redirect_uri,omitempty"`
}

var _ Data = &DataFinishRedirectURI{}

func (*DataFinishRedirectURI) Data() {}

type DataFlowReference struct {
	FlowReference FlowReference `json:"flow_reference"`
}

var _ Data = &DataFlowReference{}

func (*DataFlowReference) Data() {}

func MergeData(manyData ...Data) Data {
	m := map[string]interface{}{}

	for _, data := range manyData {
		b, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			panic(err)
		}
	}

	return mapData(m)
}
