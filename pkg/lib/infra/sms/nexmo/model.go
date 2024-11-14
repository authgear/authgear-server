package nexmo

import (
	"encoding/json"
)

// See https://developer.vonage.com/en/api/sms
type SendResponse struct {
	MessageCount string                `json:"message-count"`
	Messages     []SendResponseMessage `json:"messages"`
}

type SendResponseMessage struct {
	// https://developer.vonage.com/en/messaging/sms/guides/troubleshooting-sms#sms-api-error-codes
	Status string `json:"status"`

	// When error, the following fields are present.
	ErrorText string `json:"error-text,omitempty"`

	// When success, the following fields are present.
	To               string `json:"to,omitempty"`
	MessageID        string `json:"message-id,omitempty"`
	RemainingBalance string `json:"remaining-balance,omitempty"`
	MessagePirce     string `json:"message-price,omitempty"`
	Network          string `json:"network,omitempty"`
	ClientRef        string `json:"client-ref,omitempty"`
	AccountRef       string `json:"account-ref,omitempty"`
}

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
