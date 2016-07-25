// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package skydb

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/asset"
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

func NewEmptyRecordID() RecordID {
	return RecordID{"", ""}
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

// IsEmpty returns whether the RecordID is empty.
func (id *RecordID) IsEmpty() bool {
	return id.Type == "" && id.Key == ""
}

// RecordACLEntry grants access to a record by relation or by user_id
type RecordACLEntry struct {
	Relation string   `json:"relation,omitempty"`
	Role     string   `json:"role,omitempty"`
	Level    ACLLevel `json:"level"`
	UserID   string   `json:"user_id,omitempty"`
	Public   bool     `json:"public,omitempty"`
}

// ACLLevel represent the operation a user granted on a resource
type ACLLevel string

// ReadLevel and WriteLevel is self-explanatory
const (
	ReadLevel   ACLLevel = "read"
	WriteLevel           = "write"
	CreateLevel          = "create"
)

// NewRecordACLEntryRelation returns an ACE on relation
func NewRecordACLEntryRelation(relation string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Relation: relation,
		Level:    level,
	}
}

// NewRecordACLEntryDirect returns an ACE for a specific user
func NewRecordACLEntryDirect(userID string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Relation: "$direct",
		Level:    level,
		UserID:   userID,
	}
}

// NewRecordACLEntryRole return an ACE on role
func NewRecordACLEntryRole(role string, level ACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Role:  role,
		Level: level,
	}
}

// NewRecordACLEntryPublic return an ACE on public access
func NewRecordACLEntryPublic(level ACLLevel) RecordACLEntry {
	return RecordACLEntry{
		Public: true,
		Level:  level,
	}
}

func (ace *RecordACLEntry) Accessible(userinfo *UserInfo, level ACLLevel) bool {
	if ace.Public {
		return ace.AccessibleLevel(level)
	}
	if userinfo == nil {
		return false
	}
	if userinfo.ID == ace.UserID {
		if ace.AccessibleLevel(level) {
			return true
		}
	}
	for _, role := range userinfo.Roles {
		if role == ace.Role {
			if ace.AccessibleLevel(level) {
				return true
			}
		}
	}
	return false
}

func (ace *RecordACLEntry) AccessibleLevel(level ACLLevel) bool {
	if level == ReadLevel {
		return true
	}
	if level == ace.Level && level == WriteLevel {
		return true
	}
	return false
}

// RecordACL is a list of ACL entries defining access control for a record
type RecordACL []RecordACLEntry

// NewRecordACL returns a new RecordACL
func NewRecordACL(entries []RecordACLEntry) RecordACL {
	acl := make(RecordACL, len(entries))
	for i, v := range entries {
		acl[i] = v
	}
	return acl
}

// Accessible checks whether provided user info has certain access level
func (acl RecordACL) Accessible(userinfo *UserInfo, level ACLLevel) bool {
	if len(acl) == 0 {
		// default behavior of empty ACL
		return true
	}

	accessible := false
	for _, ace := range acl {
		if ace.Accessible(userinfo, level) {
			accessible = true
		}
	}

	return accessible
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

// NewEmptyReference returns a reference that is empty
func NewEmptyReference() Reference {
	return Reference{
		NewEmptyRecordID(),
	}
}

func (reference *Reference) Type() string {
	return reference.ID.Type
}

// IsEmpty returns whether the reference is empty.
func (reference *Reference) IsEmpty() bool {
	return reference.ID.IsEmpty()
}

// Location represent a point of geometry.
//
// It being an array of two floats is intended to provide no-copy conversion
// between paulmach/go.geo.Point.
type Location [2]float64

// NewLocation returns a new Location
func NewLocation(lng, lat float64) Location {
	return Location{lng, lat}
}

// Lng returns the longitude
func (loc Location) Lng() float64 {
	return loc[0]
}

// Lat returns the Latitude
func (loc Location) Lat() float64 {
	return loc[1]
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

func (r *Record) Accessible(userinfo *UserInfo, level ACLLevel) bool {
	if r.ACL == nil {
		return true
	}
	userID := ""
	if userinfo != nil {
		userID = userinfo.ID
	}
	if r.DatabaseID != "" && r.DatabaseID != userID {
		return false
	}
	if r.OwnerID == userID {
		return true
	}

	return r.ACL.Accessible(userinfo, level)
}

// RecordSchema is a mapping of record key to its value's data type or reference
type RecordSchema map[string]FieldType

// FieldType represents the kind of data living within a field of a RecordSchema.
type FieldType struct {
	Type          DataType
	ReferenceType string     // used only by TypeReference
	Expression    Expression // used by Computed Keys
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

// IsNumberCompatibleType returns true if the type is a numeric type
func (t DataType) IsNumberCompatibleType() bool {
	switch t {
	case TypeNumber, TypeInteger, TypeSequence:
		return true
	default:
		return false
	}
}

func (f FieldType) ToSimpleName() string {
	switch f.Type {
	case TypeString:
		return "string"
	case TypeNumber:
		return "number"
	case TypeBoolean:
		return "boolean"
	case TypeJSON:
		return "json"
	case TypeReference:
		return fmt.Sprintf("ref(%s)", f.ReferenceType)
	case TypeLocation:
		return "location"
	case TypeDateTime:
		return "datetime"
	case TypeAsset:
		return "asset"
	case TypeACL:
		return "acl"
	case TypeInteger:
		return "integer"
	case TypeSequence:
		return "sequence"
	}
	return ""
}

func SimpleNameToFieldType(s string) (result FieldType, err error) {
	switch s {
	case "string":
		result.Type = TypeString
	case "number":
		result.Type = TypeNumber
	case "boolean":
		result.Type = TypeBoolean
	case "json":
		result.Type = TypeJSON
	case "location":
		result.Type = TypeLocation
	case "datetime":
		result.Type = TypeDateTime
	case "asset":
		result.Type = TypeAsset
	case "acl":
		result.Type = TypeACL
	case "integer":
		result.Type = TypeInteger
	case "sequence":
		result.Type = TypeSequence
	default:
		if regexp.MustCompile(`^ref\(.+\)$`).MatchString(s) {
			result.Type = TypeReference
			result.ReferenceType = s[4 : len(s)-1]
		} else {
			err = fmt.Errorf("Unexpected type name: %s", s)
			return
		}
	}

	return
}
