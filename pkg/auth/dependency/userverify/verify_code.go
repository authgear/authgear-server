package userverify

import (
	"crypto/subtle"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type VerifyCode struct {
	ID         string
	UserID     string
	LoginIDKey string
	LoginID    string
	Code       string
	Consumed   bool
	CreatedAt  time.Time
}

func NewVerifyCode() VerifyCode {
	return VerifyCode{
		ID: uuid.New(),
	}
}

func (code VerifyCode) Check(inputCode string) bool {
	input := []byte(inputCode)
	expected := []byte(code.Code)

	if len(input) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare(input, expected) == 1
}
