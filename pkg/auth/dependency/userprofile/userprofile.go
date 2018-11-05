package userprofile

import (
	"encoding/json"
	"time"
)

var (
	timeNow = func() time.Time { return time.Now().UTC() }
)

// Meta is meta data part of a user profile record
type Meta struct {
	ID         string                 `json:"_id"`
	Type       string                 `json:"_type"`
	RecordID   string                 `json:"_recordID"`
	RecordType string                 `json:"_recordType"`
	Access     map[string]interface{} `json:"_access"`
	OwnerID    string                 `json:"_ownerID"`
	CreatedAt  time.Time              `json:"_createdAt"`
	CreatedBy  string                 `json:"_createdBy"`
	UpdatedAt  time.Time              `json:"_updatedAt"`
	UpdatedBy  string                 `json:"_updatedBy"`
}

// Data refers the profile info of a user,
// like username, email, age, phone number
type Data map[string]interface{}

// UserProfile refers user profile data type
type UserProfile struct {
	Meta
	Data
}

type Store interface {
	CreateUserProfile(userID string, data Data) (UserProfile, error)
	GetUserProfile(userID string) (UserProfile, error)
}

func (u UserProfile) MarshalJSON() ([]byte, error) {
	var metaJSON, _ = json.Marshal(u.Meta)
	var dataJSON, _ = json.Marshal(u.Data)
	var result map[string]interface{}
	json.Unmarshal(metaJSON, &result)
	json.Unmarshal(dataJSON, &result)
	return json.Marshal(result)
}
