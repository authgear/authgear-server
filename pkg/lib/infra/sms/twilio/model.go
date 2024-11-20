package twilio

import (
	"encoding/json"
)

// https://www.twilio.com/docs/messaging/api/message-resource#message-properties
type SendResponse struct {
	Body                *string                `json:"body,omitempty"`
	NumSegments         *string                `json:"num_segments,omitempty"`
	Direction           *string                `json:"direction,omitempty"`
	From                *string                `json:"from,omitempty"`
	To                  *string                `json:"to,omitempty"`
	DateUpdated         *string                `json:"date_updated,omitempty"`
	Price               *string                `json:"price,omitempty"`
	ErrorMessage        *string                `json:"error_message,omitempty"`
	URI                 *string                `json:"uri,omitempty"`
	AccountSID          *string                `json:"account_sid,omitempty"`
	NumMedia            *string                `json:"num_media,omitempty"`
	Status              *string                `json:"status,omitempty"`
	MessagingServiceSID *string                `json:"messaging_service_sid,omitempty"`
	SID                 *string                `json:"sid,omitempty"`
	DateSent            *string                `json:"date_sent,omitempty"`
	DateCreated         *string                `json:"date_created,omitempty"`
	ErrorCode           *int                   `json:"error_code,omitempty"`
	PriceUnit           *string                `json:"price_unit,omitempty"`
	APIVersion          *string                `json:"api_version,omitempty"`
	SubresourceURIs     map[string]interface{} `json:"subresource_uris,omitempty"`
}

func ParseSendResponse(jsonData []byte) (*SendResponse, error) {
	var response SendResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
