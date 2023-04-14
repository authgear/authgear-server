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
	Purpose  string    `json:"purpose"`
	Form     Form      `json:"form"`
	Code     string    `json:"code"`
	ExpireAt time.Time `json:"expire_at"`
	Consumed bool      `json:"consumed"`

	UserInputtedCode string `json:"user_inputted_code,omitempty"`
	AppID            string `json:"app_id,omitempty"`
	WebSessionID     string `json:"web_session_id,omitempty"`
	WorkflowID       string `json:"workflow_id,omitempty"`
}
