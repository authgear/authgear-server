package model

type SIWEVerificationRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type SIWEVerifiedData struct {
	Message          string `json:"message"`
	Signature        string `json:"signature"`
	EncodedPublicKey string `json:"encoded_public_key"`
}
