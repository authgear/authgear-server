package anonymous

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	codeAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type PromotionCode struct {
	AppID  string `json:"app_id"`
	UserID string `json:"user_id"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	CodeHash  string    `json:"code_hash"`
}

func GeneratePromotionCode() string {
	code := rand.StringWithAlphabet(32, codeAlphabet, rand.SecureRand)
	return code
}

func HashPromotionCode(code string) string {
	return crypto.SHA256String(code)
}
