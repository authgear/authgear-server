package model

import "github.com/lestrrat-go/jwx/jwk"

type SIWEVerificationRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type SIWEVerifiedData struct {
	Message          string  `json:"message"`
	Signature        string  `json:"signature"`
	EncodedPublicKey jwk.Key `json:"encoded_public_key"`
}
