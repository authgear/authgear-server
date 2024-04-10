package e2eclient

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type FlowReference struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type FlowAction struct {
	Type           string          `json:"type"`
	Identification string          `json:"identification,omitempty"`
	Authentication string          `json:"authentication,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
}

type FlowResponse struct {
	StateToken string      `json:"state_token"`
	Type       string      `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Action     *FlowAction `json:"action,omitempty"`
}

type HTTPResponse struct {
	Result *FlowResponse       `json:"result,omitempty"`
	Error  *apierrors.APIError `json:"error,omitempty"`
}
