package skydb

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/asset"
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

	level, _ := m["level"].(string)
	var entryLevel ACLLevel
	switch level {
	case "read":
		entryLevel = ReadLevel
	case "write":
		entryLevel = WriteLevel
	case "":
		return errors.New("empty level")
	default:
		return fmt.Errorf("unknown level = %s", level)
	}

	relation, _ := m["relation"].(string)
	if relation == "" {
		return errors.New("empty relation")
	}

	var userID string
	if relation == "$direct" {
		userID, _ = m["user_id"].(string)
		if userID == "" {
			return errors.New(`empty user_id when relation = "$direct"`)
		}
	}

	entry.Level = entryLevel
	entry.Relation = relation
	entry.UserID = userID

	return nil
}

// NewRecordACLEntryRelation returns an ACE on relation
func NewRecordACLEntryRelation(relation string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{relation, level, ""}
}

// NewRecordACLEntryDirect returns an ACE for a specific user
func NewRecordACLEntryDirect(userID string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{"$direct", level, userID}
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

type Asset struct {
	Name        string
	ContentType string
	Size        int64
	Public      bool
	Signer      asset.URLSigner
}

// SignedURL will try to return a signedURL with the injected Signer.
func (a *Asset) SignedURL() string {
	if a.Signer == nil {
		log.Warnf("Unable to generate signed url of asset because no singer is injected.")
		return ""
	}

	url, err := a.Signer.SignedURL(a.Name)
	if err != nil {
		log.Warnf("Unable to generate signed url: %v", err)
	}
	return url
}

type Reference struct {
	ID RecordID
}

func NewReference(recordType string, id string) Reference {
	return Reference{
		NewRecordID(recordType, id),
	}
}

func (reference *Reference) Type() string {
	return reference.ID.Type
}

// Location represent a point of geometry.
//
// It being an array of two floats is intended to provide no-copy conversion
// between paulmach/go.geo.Point.
type Location [2]float64

// NewLocation returns a new Location
func NewLocation(lng, lat float64) *Location {
	return &Location{lng, lat}
}

// Lng returns the longitude
func (loc *Location) Lng() float64 {
	return loc[0]
}

// SetLng sets the longitude
func (loc *Location) SetLng(lng float64) {
	loc[0] = lng
}

// Lat returns the Latitude
func (loc *Location) Lat() float64 {
	return loc[1]
}

// SetLat sets the Latitude
func (loc *Location) SetLat(lat float64) {
	loc[1] = lat
}

// String returns a human-readable representation of this Location.
// Coincidentally it is in WKT.
func (loc Location) String() string {
	return fmt.Sprintf("POINT(%g %g)", loc[0], loc[1])
}

// Sequence is a bogus data type for creating a sequence field
// via JIT schema migration
type Sequence struct{}

// A Data represents a key-value object used for storing ODRecord.
type Data map[string]interface{}

// Record is the primary entity of storage in Skygear.
type Record struct {
	ID         RecordID
	DatabaseID string `json:"-"`
	OwnerID    string
	CreatedAt  time.Time
	CreatorID  string
	UpdatedAt  time.Time
	UpdaterID  string
	ACL        RecordACL
	Data       Data
	Transient  Data `json:"-"`
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
		case "_created_at":
			return r.CreatedAt
		case "_created_by":
			return r.CreatorID
		case "_updated_at":
			return r.UpdatedAt
		case "_updated_by":
			return r.UpdaterID
		case "_transient":
			return r.Transient
		default:
			if strings.HasPrefix(key, "_transient_") {
				return r.Transient[strings.TrimPrefix(key, "_transient_")]
			}
			panic(fmt.Sprintf("unknown reserved key: %v", key))
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
		case "_created_at":
			r.CreatedAt = i.(time.Time)
		case "_created_by":
			r.CreatorID = i.(string)
		case "_updated_at":
			r.UpdatedAt = i.(time.Time)
		case "_updated_by":
			r.UpdaterID = i.(string)
		case "_transient":
			r.Transient = i.(Data)
		default:
			if strings.HasPrefix(key, "_transient_") {
				if r.Transient == nil {
					r.Transient = Data{}
				}
				r.Transient[strings.TrimPrefix(key, "_transient_")] = i
			} else {
				panic(fmt.Sprintf("unknown reserved key: %v", key))
			}
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
	ReferenceType string      // used only by TypeReference
	Expression    *Expression // used by Computed Keys
}

// DataType defines the type of data that can saved into an skydb database
//go:generate stringer -type=DataType
type DataType uint

// List of persistable data types in skydb
const (
	TypeString DataType = iota + 1
	TypeNumber
	TypeBoolean
	TypeJSON
	TypeReference
	TypeLocation
	TypeDateTime
	TypeAsset
	TypeACL
	TypeInteger
	TypeSequence
)
