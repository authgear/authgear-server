package userprofile

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
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
	CreatedAt  time.Time              `json:"_created_at"`
	CreatedBy  string                 `json:"_created_by"`
	UpdatedAt  time.Time              `json:"_updated_at"`
	UpdatedBy  string                 `json:"_updated_by"`
}

// Data refers the profile info of a user,
// like username, email, age, phone number
type Data map[string]interface{}

// Record refers the data type of a record
type Record map[string]interface{}

// UserProfile refers user profile data type
type UserProfile struct {
	Meta
	Data
}

type Store interface {
	CreateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (UserProfile, error)
	GetUserProfile(userID string) (UserProfile, error)
	UpdateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (UserProfile, error)
}

func (u UserProfile) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.ToMap())
}

func (u UserProfile) ToMap() map[string]interface{} {
	var metaJSON, _ = json.Marshal(u.Meta)
	var dataJSON, _ = json.Marshal(u.Data)
	var metaMap map[string]interface{}
	json.Unmarshal(metaJSON, &metaMap)
	var dataMap map[string]interface{}
	json.Unmarshal(dataJSON, &dataMap)

	// Merge meta and data into one map
	var result = metaMap
	for k, v := range dataMap {
		metaMap[k] = v
	}
	return result
}

func (u *UserProfile) UnmarshalJSON(b []byte) error {
	var record Record
	err := json.Unmarshal(b, &record)
	if err != nil {
		return err
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	err = json.Unmarshal(recordJSON, &u.Meta)
	if err != nil {
		return err
	}
	u.Data = make(map[string]interface{})
	for k, v := range record {
		if !strings.HasPrefix(k, "_") {
			u.Data[k] = v
		}
	}

	return nil
}
