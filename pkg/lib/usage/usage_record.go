package usage

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// nolint:golint
type UsageRecordName string

const (
	ActiveUser          UsageRecordName = "active-user"
	SMSSent             UsageRecordName = "sms-sent"
	EmailSent           UsageRecordName = "email-sent"
	WhatsappOTPVerified UsageRecordName = "whatsapp-otp-verified"
)

// nolint:golint
type UsageRecord struct {
	ID             string
	AppID          string
	Name           UsageRecordName
	Period         string
	StartTime      time.Time
	EndTime        time.Time
	Count          int
	AlertData      map[string]interface{}
	StripTimestamp *time.Time
}

func NewUsageRecord(appID string, name UsageRecordName, count int, period periodical.Type, startTime time.Time, endTime time.Time) *UsageRecord {
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
