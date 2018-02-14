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
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/asset"
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

// Geometry represent a geometry in GeoJSON.
type Geometry map[string]interface{}

// Sequence is a bogus data type for creating a sequence field
// via JIT schema migration
type Sequence struct{}

// Unknown is a bogus data type denoting the type of a field is unknown.
type Unknown struct {
	UnderlyingType string
}

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

// Copy makes a shadow copy of itself
func (d Data) Copy() Data {
	dataCopy := Data{}
	for key, value := range d {
		dataCopy[key] = value
	}
	return dataCopy
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

func (r *Record) Remove(key string) {
	if key[0] == '_' {
		panic(fmt.Sprintf("cannot remove reserved key: %s", key))
	}

	delete(r.Data, key)
}

func (r *Record) UserKeys() []string {
	var keys []string
	for key := range r.Data {
		keys = append(keys, key)
	}
	return keys
}

func (r *Record) Accessible(authinfo *AuthInfo, level RecordACLLevel) bool {
	if r.ACL == nil {
		return true
	}
	userID := ""
	if authinfo != nil {
		userID = authinfo.ID
	}
	if r.DatabaseID != "" && r.DatabaseID != userID {
		return false
	}
	if r.OwnerID == userID {
		return true
	}

	return r.ACL.Accessible(authinfo, level)
}

// Copy copies the content of the record.
func (r *Record) Copy() Record {
	dst := Record{}
	dst = *r

	if r.Data != nil {
		dst.Data = r.Data.Copy()
	}

	if r.Transient != nil {
		dst.Transient = r.Transient.Copy()
	}

	return dst
}

// Apply modifies the content of the record with the specified record.
func (r *Record) Apply(src *Record) {
	r.ACL = src.ACL

	if src.Data != nil {
		if r.Data == nil {
			r.Data = Data{}
		}
		for key, value := range src.Data {
			r.Data[key] = value
		}
	}

	if src.Transient != nil {
		if r.Transient == nil {
			r.Transient = Data{}
		}
		for key, value := range src.Transient {
			r.Transient[key] = value
		}
	}
}

// MergedCopy is similar to copy but the copy contains data dictionary
// which is creating by copying the original and apply the specified dictionary.
func (r *Record) MergedCopy(merge *Record) Record {
	dst := r.Copy()
	dst.Apply(merge)
	return dst
}

// Index indicates the value of fields within a record type cannot be duplicated
type Index struct {
	Fields []string
}

// RecordSchema is a mapping of record key to its value's data type or reference
type RecordSchema map[string]FieldType

// DefinitionCompatibleTo returns if a record having the specified RecordSchema
//
// can be saved to a database table of this RecordSchema.
//
// This function is not associative. In other words, `a.fn(b) != b.fn(a)`.
func (schema RecordSchema) DefinitionCompatibleTo(other RecordSchema) bool {
	if len(schema) < len(other) {
		return false
	}

	for k, myFieldType := range other {
		otherFieldType, ok := schema[k]
		if !ok {
			return false
		}

		if !myFieldType.DefinitionCompatibleTo(otherFieldType) {
			return false
		}
	}
	return true
}

func (schema RecordSchema) HasField(field string) bool {
	_, found := schema[field]
	return found
}

func (schema RecordSchema) HasFields(fields []string) bool {
	for _, field := range fields {
		found := schema.HasField(field)
		if !found {
			return false
		}
	}

	return true
}

// FieldType represents the kind of data living within a field of a RecordSchema.
type FieldType struct {
	Type           DataType
	ReferenceType  string     // used only by TypeReference
	Expression     Expression // used by Computed Keys
	UnderlyingType string     // indicates the underlying (pq) type
}

// DefinitionCompatibleTo returns if a value of the specified FieldType can
// be saved to a database column of this FieldType.
//
// When a FieldType is compatible with another FieldType, it also means
// it is possible to cast value of a type to another type. Whether the cast
// is successful is subject to the actual value, whether it will
// lose number precision for example.
//
// This function is not associative. In other words, `a.fn(b) != b.fn(a)`.
func (f FieldType) DefinitionCompatibleTo(other FieldType) bool {
	if f.Type == TypeReference {
		return f.Type == other.Type && f.ReferenceType == other.ReferenceType
	}

	if f.Type.IsNumberCompatibleType() && other.Type.IsNumberCompatibleType() {
		return true
	}

	if f.Type == TypeGeometry && other.Type.IsGeometryCompatibleType() {
		// Note: Saving skydb.Location to skydb.Geometry is currently
		// not supported (see #343)
		//return true
		return other.Type == TypeGeometry
	}

	return f.Type == other.Type
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
	case TypeGeometry:
		return "geometry"
	case TypeUnknown:
		return "unknown"
	}
	return ""
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
	TypeGeometry
	TypeUnknown
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

func (t DataType) IsGeometryCompatibleType() bool {
	return t == TypeLocation || t == TypeGeometry
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
	case "geometry":
		result.Type = TypeGeometry
	case "unknown":
		result.Type = TypeUnknown
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

// DeriveFieldType finds the FieldType of the specified value.
// nolint: gocyclo
func DeriveFieldType(value interface{}) (fieldType FieldType, err error) {
	switch val := value.(type) {
	default:
		kind := reflect.ValueOf(val).Kind()
		if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
			fieldType = FieldType{
				Type: TypeJSON,
			}
		} else {
			err = fmt.Errorf("got unrecognized type = %T", value)
		}
	case nil:
		err = errors.New("cannot derive field type from nil")
	case int64:
		fieldType = FieldType{
			Type: TypeInteger,
		}
	case float64:
		fieldType = FieldType{
			Type: TypeNumber,
		}
	case string:
		fieldType = FieldType{
			Type: TypeString,
		}
	case time.Time:
		fieldType = FieldType{
			Type: TypeDateTime,
		}
	case bool:
		fieldType = FieldType{
			Type: TypeBoolean,
		}
	case *Asset:
		fieldType = FieldType{
			Type: TypeAsset,
		}
	case Reference:
		v := value.(Reference)
		fieldType = FieldType{
			Type:          TypeReference,
			ReferenceType: v.Type(),
		}
	case Location:
		fieldType = FieldType{
			Type: TypeLocation,
		}
	case Sequence:
		fieldType = FieldType{
			Type: TypeSequence,
		}
	case Geometry:
		fieldType = FieldType{
			Type: TypeGeometry,
		}
	case Unknown:
		fieldType = FieldType{
			Type:           TypeUnknown,
			UnderlyingType: val.UnderlyingType,
		}
	}
	return
}
