package audit

import (
	"time"
)

type Log struct {
	ID           string                 `json:"id"`
	CreatedAt    time.Time              `json:"createdAt"`
	ActivityType string                 `json:"activityType"`
	UserID       string                 `json:"userID,omitempty"`
	IPAddress    string                 `json:"ipAddress,omitempty"`
	UserAgent    string                 `json:"userAgent,omitempty"`
	ClientID     string                 `json:"clientID,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}
