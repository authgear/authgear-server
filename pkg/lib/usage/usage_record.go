package usage

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type RecordName string

const (
	RecordNameActiveUser          RecordName = "active-user"
	RecordNameSMSSentNorthAmerica RecordName = "sms-sent.north-america"
	RecordNameSMSSentOtherRegions RecordName = "sms-sent.other-regions"
	RecordNameSMSSentTotal        RecordName = "sms-sent.total"
	RecordNameEmailSent           RecordName = "email-sent"
	RecordNameWhatsappOTPVerified RecordName = "whatsapp-otp-verified"
)

type RecordType string

const (
	RecordTypeActiveUser          RecordType = "active-user"
	RecordTypeSMSSent             RecordType = "sms-sent"
	RecordTypeEmailSent           RecordType = "email-sent"
	RecordTypeWhatsappOTPVerified RecordType = "whatsapp-otp-verified"
)

// nolint:golint
type UsageRecord struct {
	ID              string
	AppID           string
	Name            RecordName
	Period          string
	StartTime       time.Time
	EndTime         time.Time
	Count           int
	AlertData       map[string]interface{}
	StripeTimestamp *time.Time
}

func NewUsageRecord(appID string, name RecordName, count int, period periodical.Type, startTime time.Time, endTime time.Time) *UsageRecord {
	return &UsageRecord{
		ID:        uuid.New(),
		AppID:     appID,
		Name:      name,
		Count:     count,
		Period:    string(period),
		StartTime: startTime,
		EndTime:   endTime,
	}
}
