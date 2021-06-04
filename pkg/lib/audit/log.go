package audit

import (
	"encoding/json"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
)

// Log represents an audit log entry.
// The keys in JSON struct tags are in camel case
// because this struct is directly returned in the GraphQL endpoint.
// Making the keys in camel case saves us from writing boilerplate resolver code.
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

func NewLog(e *event.Event) (*Log, error) {
	var userID string
	if e.Context.UserID != nil {
		userID = *e.Context.UserID
	}

	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &Log{
		ID:           e.ID,
		CreatedAt:    time.Unix(e.Context.Timestamp, 0).UTC(),
		UserID:       userID,
		ActivityType: string(e.Type),
		IPAddress:    e.Context.IPAddress,
		UserAgent:    e.Context.UserAgent,
		ClientID:     e.Context.ClientID,
		Data:         data,
	}, nil
}
