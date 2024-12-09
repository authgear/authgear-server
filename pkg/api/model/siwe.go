package model

type SIWEPublicKey string

type SIWEVerifiedData struct {
	Message          string        `json:"message"`
	Signature        string        `json:"signature"`
	EncodedPublicKey SIWEPublicKey `json:"encoded_public_key"`
}
