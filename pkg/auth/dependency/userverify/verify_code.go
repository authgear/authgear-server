package userverify

import (
	"crypto/subtle"
	"github.com/skygeario/skygear-server/pkg/core/base32"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type VerifyCode struct {
	ID         string
	UserID     string
	LoginIDKey string
	LoginID    string
	Code       string // code alphabet must be subset of base32 alphabet
	Consumed   bool
	CreatedAt  time.Time
}

func NewVerifyCode() VerifyCode {
	return VerifyCode{
		ID: uuid.New(),
	}
}

func (code VerifyCode) Check(inputCode string) bool {
	normalizedInputCode, err := base32.Normalize(inputCode)
	if err != nil {
		return false
	}

	input := []byte(normalizedInputCode)
	expected := []byte(code.Code)

	return subtle.ConstantTimeCompare(input, expected) == 1
}
