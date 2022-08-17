package model

type SIWEVerificationRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}
