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

package recordconv

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

// MapFrom tries to map a map to a FromMapper
func MapFrom(i interface{}, fromMapper FromMapper) error {
	if m, ok := i.(map[string]interface{}); ok {
		return fromMapper.FromMap(m)
	}

	return fmt.Errorf("want map, got type = %T", i)
}

// FromMapper defines whether a type can be converted from a map
type FromMapper interface {
	FromMap(m map[string]interface{}) error
}

// ToMapper defines whether a type can be converted to a map
type ToMapper interface {
	ToMap(m map[string]interface{})
}

// ToMap converts a ToMapper to map and returns it
func ToMap(mapper ToMapper) map[string]interface{} {
	mm := map[string]interface{}{}
	mapper.ToMap(mm)
	return mm
}

// MapData is record data that can be converted from a map
type MapData map[string]interface{}

// FromMap implements FromMapper
func (data *MapData) FromMap(m map[string]interface{}) (err error) {
	var walkedData map[string]interface{}
	walkedData, err = walkData(m)
	if err != nil {
		return
	}

	*data = walkedData
	return nil
}

// ToMap implements ToMapper
func (data MapData) ToMap(m map[string]interface{}) {
	for key, value := range data {
		if mapper, ok := value.(ToMapper); ok {
			mm := map[string]interface{}{}
			mapper.ToMap(mm)
			m[key] = mm
		} else {
			m[key] = value
		}
	}
}

// MapTime is time.Time that can be converted from and to a map.
type MapTime time.Time

// FromMap implements FromMapper
func (t *MapTime) FromMap(m map[string]interface{}) error {
	datei, ok := m["$date"]
	if !ok {
		return errors.New("missing compulsory field $date")
	}
	dateStr, ok := datei.(string)
	if !ok {
		return fmt.Errorf("got type($date) = %T, want string", datei)
	}
	dt, err := time.Parse(time.RFC3339Nano, dateStr)
	if err != nil {
		return fmt.Errorf("failed to parse $date = %#v", dateStr)
	}

	*(*time.Time)(t) = dt.In(time.UTC)
	return nil
}

// ToMap implements ToMapper
func (t MapTime) ToMap(m map[string]interface{}) {
	m["$type"] = "date"
	m["$date"] = time.Time(t)
}

// MapAsset is record.Asset that can be converted from and to a map.
type MapAsset record.Asset

// FromMap implements FromMapper
func (asset *MapAsset) FromMap(m map[string]interface{}) error {
	namei, ok := m["$name"]
	if !ok {
		return errors.New("missing compulsory field $name")
	}
	name, ok := namei.(string)
	if !ok {
		return fmt.Errorf("got type($name) = %T, want string", namei)
	}
	if name == "" {
		return errors.New("asset's $name should not be empty")
	}
	asset.Name = name

	contentTypei, ok := m["$content_type"]
	if ok {
		contentType, ok := contentTypei.(string)
		if !ok {
			return fmt.Errorf("got type($contentType) = %T, want string", contentTypei)
		}
		asset.ContentType = contentType
	}

	return nil
}

// ToMap implements ToMapper
func (asset *MapAsset) ToMap(m map[string]interface{}) {
	m["$type"] = "asset"
	m["$name"] = asset.Name
	if asset.ContentType != "" {
		m["$content_type"] = asset.ContentType
	}
	url := (*record.Asset)(asset).SignedURL()
	if url != "" {
		m["$url"] = url
	}
}

// MapReference is record.Reference that can be converted from and to a map.
type MapReference record.Reference

// FromMap implements FromMapper
func (ref *MapReference) FromMap(m map[string]interface{}) error {
	if recordType, ok := m["$recordType"].(string); ok {
		ref.ID.Type = recordType
	}
	if recordID, ok := m["$recordID"].(string); ok {
		ref.ID.Key = recordID
	}

	if ref.ID.Type == "" || ref.ID.Key == "" {
		// NOTE(cheungpat): Handling for deprecated fields.
		if deprecatedID, ok := m["$id"].(string); ok {
			ss := strings.SplitN(deprecatedID, "/", 2)
			if len(ss) == 1 {
				return fmt.Errorf(`ref: "_id" should be of format '{type}/{id}', got %#v`, deprecatedID)
			}
			ref.ID.Type = ss[0]
			ref.ID.Key = ss[1]
		}
	}

	if ref.ID.Type == "" {
		return errors.New("missing $recordType, expecting string")
	}

	if ref.ID.Key == "" {
		return errors.New("missing $recordID, expecting string")
	}

	return nil
}

// ToMap implements ToMapper
func (ref MapReference) ToMap(m map[string]interface{}) {
	m["$type"] = "ref"
	m["$id"] = ref.ID // NOTE(cheungpat): Fields to be deprecated.
	m["$recordID"] = ref.ID.Key
	m["$recordType"] = ref.ID.Type
}

// MapLocation is record.Location that can be converted from and to a map.
type MapLocation record.Location

// FromMap implements FromMapper
func (loc *MapLocation) FromMap(m map[string]interface{}) error {
	getFloat := func(m map[string]interface{}, key string) (float64, error) {
		i, ok := m[key]
		if !ok {
			return 0, fmt.Errorf("missing compulsory field %s", key)
		}

		f, ok := i.(float64)
		if !ok {
			return 0, fmt.Errorf("got type(%s) = %T, want number", key, i)
		}

		return f, nil
	}

	lng, err := getFloat(m, "$lng")
	if err != nil {
		return err
	}

	lat, err := getFloat(m, "$lat")
	if err != nil {
		return err
	}

	*loc = MapLocation{lng, lat}
	return nil
}

// ToMap implements ToMapper
func (loc MapLocation) ToMap(m map[string]interface{}) {
	m["$type"] = "geo"
	m["$lng"] = loc[0]
	m["$lat"] = loc[1]
}

// MapGeometry is record.Geometry that can be converted from and to a map.
type MapGeometry record.Geometry

// ToMap implements ToMapper
func (geom MapGeometry) ToMap(m map[string]interface{}) {
	m["$type"] = "geojson"
	m["$val"] = geom
}

// FromMap implements FromMapper
func (geom *MapGeometry) FromMap(m map[string]interface{}) error {
	var ok bool

	*geom, ok = m["$val"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("$val is not a map")
	}

	return nil
}

func walkData(m map[string]interface{}) (mapReturned map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	return walkMap(m, ParseLiteral), err
}

// MapKeyPath is string keypath that can be converted from a map
type MapKeyPath string

// FromMap implements FromMapper
func (p *MapKeyPath) FromMap(m map[string]interface{}) error {
	keyPath, _ := m["$val"].(string)
	if keyPath == "" {
		return errors.New("empty key path")
	}

	*p = MapKeyPath(keyPath)
	return nil
}

// ToMap implements ToMapper
func (p MapKeyPath) ToMap(m map[string]interface{}) {
	m["$type"] = "keypath"
	m["$val"] = string(p)
}

// MapRelation is a type specifying a relation between two users, but do not conform to any actual struct in record.
type MapRelation struct {
	Name      string
	Direction string
}

// FromMap implements FromMapper
func (rel *MapRelation) FromMap(m map[string]interface{}) error {
	name, _ := m["$name"].(string)
	if name == "" {
		return errors.New("empty relation name")
	}

	direction, _ := m["$direction"].(string)
	if direction == "" {
		return errors.New("empty direction")
	}

	*rel = MapRelation{name, direction}
	return nil
}

// ToMap implements ToMapper
func (rel *MapRelation) ToMap(m map[string]interface{}) {
	m["$type"] = "relation"
	m["$name"] = rel.Name
	m["$direction"] = rel.Direction
}

// MapSequence is record.Sequence that can convert to map
type MapSequence struct{}

// ToMap implements ToMapper
func (seq MapSequence) ToMap(m map[string]interface{}) {
	m["$type"] = "seq"
}

// MapUnknown is record.Unknown that can convert to map
type MapUnknown record.Unknown

// FromMap implements FromMapper
func (val *MapUnknown) FromMap(m map[string]interface{}) error {
	underlyingType, _ := m["$underlying_type"].(string)
	*val = MapUnknown{underlyingType}
	return nil
}

// ToMap implements ToMapper
func (val MapUnknown) ToMap(m map[string]interface{}) {
	m["$type"] = "unknown"
	m["$underlying_type"] = val.UnderlyingType
}

type MapACLEntry record.ACLEntry

// FromMap initializes a RecordACLEntry from a unmarshalled JSON of
// access control definition
func (ace *MapACLEntry) FromMap(m map[string]interface{}) error {
	level, _ := m["level"].(string)
	var entryLevel record.ACLLevel
	switch level {
	case "read":
		entryLevel = record.ReadLevel
	case "write":
		entryLevel = record.WriteLevel
	case "":
		return errors.New("empty level")
	default:
		return fmt.Errorf("unknown level = %s", level)
	}

	relation, hasRelation := m["relation"].(string)
	userID, hasUserID := m["user_id"].(string)
	role, hasRole := m["role"].(string)
	public, hasPublic := m["public"].(bool)
	if !hasRelation && !hasUserID && !hasRole && !hasPublic {
		return errors.New("ACLEntry must have relation, user_id, role or public")
	}

	ace.Level = entryLevel
	if hasRelation {
		ace.Relation = relation
	}
	if hasRole {
		ace.Role = role
	}
	if hasUserID {
		ace.UserID = userID
	}
	if hasPublic {
		ace.Public = public
	}
	return nil
}

type MapFieldACLEntry record.FieldACLEntry

// FromMap initializes a FieldACLEntry from a unmarshalled JSON of
// field access control definition
func (ace *MapFieldACLEntry) FromMap(m map[string]interface{}) error {
	if recordType, ok := m["record_type"].(string); ok {
		if recordType == "" {
			return fmt.Errorf(`invalid record_type "%s"`, recordType)
		}
		ace.RecordType = recordType
	} else {
		return fmt.Errorf("missing or invalid record_type")
	}

	if recordField, ok := m["record_field"].(string); ok {
		ace.RecordField = recordField
		if recordField == "" {
			return fmt.Errorf(`invalid record_field "%s"`, recordField)
		}
	} else {
		return fmt.Errorf("missing or invalid record_field")
	}

	if userRole, ok := m["user_role"].(string); ok {
		var err error
		ace.UserRole, err = record.ParseFieldUserRole(userRole)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing or invalid user_role")
	}

	if readable, ok := m["readable"].(bool); ok {
		ace.Readable = readable
	} else {
		return fmt.Errorf("missing or invalid readable")
	}

	if writable, ok := m["writable"].(bool); ok {
		ace.Writable = writable
	} else {
		return fmt.Errorf("missing or invalid writable")
	}

	if comparable, ok := m["comparable"].(bool); ok {
		ace.Comparable = comparable
	} else {
		return fmt.Errorf("missing or invalid comparable")
	}

	if discoverable, ok := m["discoverable"].(bool); ok {
		ace.Discoverable = discoverable
	} else {
		return fmt.Errorf("missing or invalid discoverable")
	}
	return nil
}

// MapWrappedRecord is record.Record that can be converted from and to a map.
type MapWrappedRecord record.Record

// FromMap implements FromMapper
func (t *MapWrappedRecord) FromMap(m map[string]interface{}) error {
	recordi, ok := m["$record"]
	if !ok {
		return errors.New("missing compulsory field $record")
	}

	recordm, ok := recordi.(map[string]interface{})
	if !ok {
		return errors.New("$record is not a map")
	}

	//	tt := (*record.Record)(t)
	return (*JSONRecord)(t).FromMap(recordm)
}

// ToMap implements ToMapper
func (t *MapWrappedRecord) ToMap(m map[string]interface{}) {
	m["$type"] = "record"
	mm := map[string]interface{}{}
	(*JSONRecord)(t).ToMap(mm)
	m["$record"] = mm
}

func walkMap(m map[string]interface{}, fn func(interface{}) interface{}) map[string]interface{} {
	for key, value := range m {
		m[key] = fn(value)
	}

	return m
}

func walkSlice(items []interface{}, fn func(interface{}) interface{}) []interface{} {
	for i, item := range items {
		items[i] = fn(item)
	}

	return items
}

// TryParseLiteral deduces whether i is a skydb data value and returns a
// parsed value.
// nolint: gocyclo
func TryParseLiteral(i interface{}) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	out = ParseLiteral(i)
	return
}

// ParseLiteral deduces whether i is a skydb data value and returns a
// parsed value.
// nolint: gocyclo
func ParseLiteral(i interface{}) interface{} {
	switch value := i.(type) {
	default:
		// considered a bug if this line is reached
		panic(fmt.Errorf("unsupported value = %T", value))
	case nil, bool, float64, string:
		// the set of value that json unmarshaller returns
		// http://golang.org/pkg/encoding/json/#Unmarshal
		return value
	case map[string]interface{}:
		kindi, typed := value["$type"]
		if !typed {
			// regular dictionary, go deeper
			return walkMap(value, ParseLiteral)
		}

		kind, ok := kindi.(string)
		if !ok {
			panic(fmt.Errorf(`got "$type"'s type = %T, want string`, kindi))
		}

		switch kind {
		case "keypath":
			var keyPath string
			mapFromOrPanic((*MapKeyPath)(&keyPath), value)
			return keyPath
		case "blob":
			panic(fmt.Errorf("unimplemented $type = %s", kind))
		case "asset":
			var asset record.Asset
			mapFromOrPanic((*MapAsset)(&asset), value)
			return &asset
		case "ref":
			var ref record.Reference
			mapFromOrPanic((*MapReference)(&ref), value)
			return ref
		case "date":
			var t time.Time
			mapFromOrPanic((*MapTime)(&t), value)
			return t
		case "geo":
			var loc record.Location
			mapFromOrPanic((*MapLocation)(&loc), value)
			return loc
		case "geojson":
			var geom record.Geometry
			mapFromOrPanic((*MapGeometry)(&geom), value)
			return geom
		case "seq":
			return record.Sequence{}
		case "unknown":
			var val record.Unknown
			mapFromOrPanic((*MapUnknown)(&val), value)
			return val
		case "relation":
			var rel MapRelation
			mapFromOrPanic((*MapRelation)(&rel), value)
			return &rel
		case "record":
			var record record.Record
			mapFromOrPanic((*MapWrappedRecord)(&record), value)
			return &record
		default:
			panic(fmt.Errorf("unknown $type = %s", kind))
		}
	case []interface{}:
		return walkSlice(value, ParseLiteral)
	}
}

// ToLiteral converts a primitive type or a skydb type into a map.
// This is the opposite of ParseLiteral.
// nolint: gocyclo
func ToLiteral(i interface{}) interface{} {
	switch value := i.(type) {
	default:
		// considered a bug if this line is reached
		panic(fmt.Errorf("unsupported value = %T", value))
	case nil, bool, float64, string, int, int64:
		return value
	case map[string]interface{}:
		return walkMap(value, ToLiteral)
	case []interface{}:
		return walkSlice(value, ToLiteral)
	case *record.Asset:
		return ToMap((*MapAsset)(value))
	case record.Reference:
		return ToMap((MapReference)(value))
	case time.Time:
		return ToMap((MapTime)(value))
	case record.Location:
		return ToMap((MapLocation)(value))
	case *record.Location:
		return ToMap((*MapLocation)(value))
	case record.Geometry:
		return ToMap((MapGeometry)(value))
	case record.Sequence:
		return ToMap((MapSequence)(value))
	case record.Unknown:
		return ToMap((MapUnknown)(value))
	case record.Record:
		return ToMap((*MapWrappedRecord)(&value))
	case *record.Record:
		return ToMap((*MapWrappedRecord)(value))
	case JSONRecord:
		return ToMap(&value)
	case *JSONRecord:
		return ToMap(value)
	}
}

func mapFromOrPanic(fromMapper FromMapper, m map[string]interface{}) {
	if err := fromMapper.FromMap(m); err != nil {
		panic(err)
	}
}
