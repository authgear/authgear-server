package oddb

import (
	"bytes"
	"errors"
	"fmt"
)

// RecordID identifies an unique record in a Database
type RecordID struct {
	Type string
	Key  string
}

// NewRecordID returns a new RecordID
func NewRecordID(recordType string, id string) RecordID {
	return RecordID{recordType, id}
}

// String implements the fmt.Stringer interface.
func (id RecordID) String() string {
	return id.Type + "/" + id.Key
}

// MarshalText implements the encoding.TextUnmarshaler interface.
func (id RecordID) MarshalText() ([]byte, error) {
	return []byte(id.Type + "/" + id.Key), nil
}

// UnmarshalText implements the encoding.TextMarshaler interface.
func (id *RecordID) UnmarshalText(data []byte) error {
	splited := bytes.SplitN(data, []byte("/"), 2)

	if len(splited) < 2 {
		return errors.New("invalid record id")
	}

	id.Type = string(splited[0])
	id.Key = string(splited[1])

	return nil
}

// A Data represents a key-value object used for storing ODRecord.
type Data map[string]interface{}

// Record is the primary entity of storage in Ourd.
type Record struct {
	ID     RecordID `json:"_id"`
	Data   Data     `json:"data"`
	UserID string   `json:"-"` // the owner's ID
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
			return r.ID.Type
		case "_id":
			return r.ID.Key
		case "_user_id":
			return r.UserID
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
			r.ID.Type = i.(string)
		case "_id":
			r.ID.Key = i.(string)
		case "_user_id":
			r.UserID = i.(string)
		default:
			panic(fmt.Sprintf("unknown reserved key: %v", key))
		}
	} else {
		r.Data[key] = i
	}
}

// DataType defines the type of data that can saved into an oddb database
//go:generate stringer -type=DataType
type DataType uint

// List of persistable data types in oddb
const (
	TypeString DataType = iota + 1
	TypeNumber
	TypeBoolean
	TypeJSON
	TypeReference
	TypeLocation // not implemented
	TypeDateTime
	TypeData // not implemented
)

type Schema struct {
	Type DataType
}

// RecordSchema is a mapping of record key to its value's data type or reference
type RecordSchema map[string]Schema
