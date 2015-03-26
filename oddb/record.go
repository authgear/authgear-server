package oddb

import (
	"encoding/json"
	"fmt"
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

// Get returns the value specified by key. If no value is associated
// with the specified key, it returns nil.
//
// Get also supports getting reserved fields starting with "_". If such
// reserved field does not exists, it returns nil.
func (r *Record) Get(key string) interface{} {
	if key[0] == '_' {
		switch key {
		case "_type":
			return r.Type
		case "_id":
			return r.Key
		default:
			return nil
		}
	} else {
		return r.Data[key]
	}
}

// Set associates key with the value i in this record.
//
// Set is able to associate reserved key name starting with "_" as well.
// If there is no such key, it panics.
func (r *Record) Set(key string, i interface{}) {
	if key[0] == '_' {
		switch key {
		case "_type":
			r.Type = i.(string)
		case "_id":
			r.Key = i.(string)
		default:
			panic(fmt.Sprintf("unknown reserved key: %v", key))
		}
	} else {
		r.Data[key] = i
	}
}

type Ref struct {
	Key  string
	Type string
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
