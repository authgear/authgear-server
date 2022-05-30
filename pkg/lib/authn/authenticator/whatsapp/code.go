package whatsapp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	WhatsappCodeDuration = duration.UserInteraction
)

type Code struct {
	AppID            string    `json:"app_id"`
	WebSessionID     string    `json:"web_session_id"`
	Phone            string    `json:"phone"`
	Code             string    `json:"code"`
	UserInputtedCode string    `json:"user_inputted_code"`
	ExpireAt         time.Time `json:"expire_at"`
}
