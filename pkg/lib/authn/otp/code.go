package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	WhatsappCodeDuration = duration.UserInteraction
)

type Code struct {
	Target   string    `json:"target"`
	Code     string    `json:"code"`
	ExpireAt time.Time `json:"expire_at"`

	UserInputtedCode string `json:"user_inputted_code,omitempty"`
	AppID            string `json:"app_id,omitempty"`
	WebSessionID     string `json:"web_session_id,omitempty"`
	WorkflowID       string `json:"workflow_id,omitempty"`
}
