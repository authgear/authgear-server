package oddb

import (
	"bytes"
	"encoding/json"
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

// RecordACLEntry grants access to a record by relation or by user_id
type RecordACLEntry struct {
	Relation string   `json:"relation"`
	Level    ACLLevel `json:"level"`
	UserID   string   `json:"user_id,omitempty"`
}

// ACLLevel represent the operation a user granted on a resource
type ACLLevel string

// ReadLevel and WriteLevel is self-explanatory
const (
	ReadLevel  ACLLevel = "read"
	WriteLevel          = "write"
)

// InitFromJSON initializes a RecordACLEntry from a unmarshalled JSON of
// access control definition
func (entry *RecordACLEntry) InitFromJSON(i interface{}) error {
	m, ok := i.(map[string]interface{})
	if !ok {
		return fmt.Errorf("want a dictionary, got a %T", i)
	}

	entry.Relation, _ = m["relation"].(string)
	if entry.Relation == "" {
		return errors.New("missing relation field")
	}

	level, _ := m["level"].(string)
	switch level {
	case "read":
		entry.Level = ReadLevel
	case "write":
		entry.Level = WriteLevel
	default:
		return errors.New("missing level field")
	}

	entry.UserID, _ = m["user_id"].(string)
	if entry.Relation == "" {
		return errors.New("missing user_id field")
	}

	return nil
}

// RecordACLEntry returns an ACE on relation
func NewRecordACLEntryRelation(relation string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{relation, level, ""}
}

// RecordACLEntry returns an ACE for a specific user
func NewRecordACLEntryDirect(user_id string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{"$direct", level, user_id}
}

// RecordACL is a list of ACL entries defining access control for a record
type RecordACL []RecordACLEntry

// InitFromJSON initializes a RecordACL
func (acl *RecordACL) InitFromJSON(i interface{}) error {
	if i == nil {
		*acl = nil
		return nil
	}

	l, ok := i.([]interface{})
	if !ok {
		return fmt.Errorf("want an array, got %T", i)
	}

	for i, v := range l {
		entry := RecordACLEntry{}
		if err := entry.InitFromJSON(v); err != nil {
			return fmt.Errorf(`invalid access entry at %d: %v`, i, err)
		}
		entries := (*[]RecordACLEntry)(acl)
		*entries = append(*entries, entry)
	}

	return nil
}

// NewRecordACL returns a new RecordACL
func NewRecordACL(entries []RecordACLEntry) RecordACL {
	acl := make(RecordACL, len(entries))
	for i, v := range entries {
		acl[i] = v
	}
	return acl
}

type Reference struct {
	ID RecordID `json:"_id"`
}

func (ref Reference) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string   `json:"$type"`
		ID   RecordID `json:"$id"`
	}{
		"ref",
		ref.ID,
	})
}

func NewReference(recordType string, id string) Reference {
	return Reference{
		NewRecordID(recordType, id),
	}
}

func (reference *Reference) Type() string {
	return reference.ID.Type
}

// A Data represents a key-value object used for storing ODRecord.
type Data map[string]interface{}

// Record is the primary entity of storage in Ourd.
type Record struct {
	ID         RecordID  `json:"_id"`
	Data       Data      `json:"data"`
	DatabaseID string    `json:"-"` // empty for public database
	OwnerID    string    `json:"_ownerID,omitempty"`
	ACL        RecordACL `json:"_access"`
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
		case "_database_id":
			return r.DatabaseID
		case "_owner_id":
			return r.OwnerID
		case "_access":
			return r.ACL
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
		case "_database_id":
			r.DatabaseID = i.(string)
		case "_owner_id":
			r.OwnerID = i.(string)
		case "_access":
			r.ACL = i.(RecordACL)
		default:
			panic(fmt.Sprintf("unknown reserved key: %v", key))
		}
	} else {
		r.Data[key] = i
	}
}

// RecordSchema is a mapping of record key to its value's data type or reference
type RecordSchema map[string]FieldType

// FieldType represents the kind of data living within a field of a RecordSchema.
type FieldType struct {
	Type          DataType
	ReferenceType string // used only by TypeReference
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
	TypeACL
)
