package userverify

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type VerifyCode struct {
	ID          string
	UserID      string
	RecordKey   string
	RecordValue string
	Code        string
	Consumed    bool
	CreatedAt   time.Time

	// cached value, by computing CreatedAt + expiry
	expireAt *time.Time
}

func NewVerifyCode() VerifyCode {
	return VerifyCode{
		ID: uuid.New(),
	}
}

func (v *VerifyCode) ExpireAt() *time.Time {
	return v.expireAt
}
