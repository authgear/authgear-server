package oddb

import (
	"encoding/json"
	"time"
)

// A Data represents a key-value object used for storing ODRecord.
type Data map[string]interface{}

// Record is the primary entity of storage in Ourd.
type Record struct {
	Type string `json:"_type"`
	Key  string `json:"_id"`
	Data `json:"data"`
}

// A Datetime represent an instance in time.
// Internally it is an alias of time.Time with custom (Un)Marshalling Logic.
type Datetime time.Time

type transportDatetime struct {
	Type     string `json:"$type"`
	Datetime `json:"$date"`
}

// MarshalJSON implements the json.Marshaler interface.
func (dt Datetime) MarshalJSON() ([]byte, error) {
	return json.Marshal(transportDatetime{
		Type:     "date",
		Datetime: dt,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (dt *Datetime) UnmarshalJSON(data []byte) (err error) {
	tdt := transportDatetime{}
	if err := json.Unmarshal(data, &tdt); err != nil {
		return err
	}
	*dt = tdt.Datetime
	return nil
}
