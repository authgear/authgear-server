package otp

import (
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	WhatsappCodeDuration = duration.UserInteraction
)

type Code struct {
	Target   string    `json:"target"`
	Purpose  Purpose   `json:"purpose"`
	Form     Form      `json:"form"`
	Code     string    `json:"code"`
	ExpireAt time.Time `json:"expire_at"`
	Consumed bool      `json:"consumed"`

	UserInputtedCode string `json:"user_inputted_code,omitempty"`
	UserID           string `json:"user_id,omitempty"`

	// For interaction
	WebSessionID string `json:"web_session_id,omitempty"`

	// For legacy workflow
	WorkflowID string `json:"workflow_id,omitempty"`

	// For authentication flow
	AuthenticationFlowWebsocketChannelName string        `json:"authentication_flow_websocket_channel_name,omitempty"`
	AuthenticationFlowType                 string        `json:"authentication_flow_type,omitempty"`
	AuthenticationFlowName                 string        `json:"authentication_flow_name,omitempty"`
	AuthenticationFlowJSONPointer          jsonpointer.T `json:"authentication_flow_json_pointer,omitempty"`
}
