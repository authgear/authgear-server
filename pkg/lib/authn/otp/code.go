package otp

import (
	"time"
)

type Code struct {
	Code             string    `json:"code"`
	ExpireAt         time.Time `json:"expire_at"`
	UserInputtedCode string    `json:"user_inputted_code"`
}
