package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	nonceAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateOpenIDConnectNonce() string {
	nonce := rand.StringWithAlphabet(32, nonceAlphabet, rand.SecureRand)
	return nonce
}
