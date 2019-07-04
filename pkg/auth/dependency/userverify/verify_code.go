package userverify

import (
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
