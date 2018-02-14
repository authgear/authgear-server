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

package skyconv

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
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

// MapAsset is skydb.Asset that can be converted from and to a map.
type MapAsset skydb.Asset

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
	m["$content_type"] = asset.ContentType
	url := (*skydb.Asset)(asset).SignedURL()
	if url != "" {
		m["$url"] = url
	}
}

// MapReference is skydb.Reference that can be converted from and to a map.
type MapReference skydb.Reference

// FromMap implements FromMapper
func (ref *MapReference) FromMap(m map[string]interface{}) error {
	idi, ok := m["$id"]
	if !ok {
		return errors.New("referencing without $id")
	}
	id, ok := idi.(string)
	if !ok {
		return fmt.Errorf("got reference type($id) = %T, want string", idi)
	}
	ss := strings.SplitN(id, "/", 2)
	if len(ss) == 1 {
		return fmt.Errorf(`ref: "_id" should be of format '{type}/{id}', got %#v`, id)
	}

	ref.ID.Type = ss[0]
	ref.ID.Key = ss[1]
	return nil
}

// ToMap implements ToMapper
func (ref MapReference) ToMap(m map[string]interface{}) {
	m["$type"] = "ref"
	m["$id"] = ref.ID
}

// MapLocation is skydb.Location that can be converted from and to a map.
type MapLocation skydb.Location

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

// MapGeometry is skydb.Geometry that can be converted from and to a map.
type MapGeometry skydb.Geometry

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

	return walkMap(m), err
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

// MapRelation is a type specifying a relation between two users, but do not conform to any actual struct in skydb.
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

// MapSequence is skydb.Sequence that can convert to map
type MapSequence struct{}

// ToMap implements ToMapper
func (seq MapSequence) ToMap(m map[string]interface{}) {
	m["$type"] = "seq"
}

// MapUnknown is skydb.Unknown that can convert to map
type MapUnknown skydb.Unknown

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

type MapACLEntry skydb.RecordACLEntry

// FromMap initializes a RecordACLEntry from a unmarshalled JSON of
// access control definition
func (ace *MapACLEntry) FromMap(m map[string]interface{}) error {
	level, _ := m["level"].(string)
	var entryLevel skydb.RecordACLLevel
	switch level {
	case "read":
		entryLevel = skydb.ReadLevel
	case "write":
		entryLevel = skydb.WriteLevel
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

type MapFieldACLEntry skydb.FieldACLEntry

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
		ace.UserRole, err = skydb.ParseFieldUserRole(userRole)
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

func walkMap(m map[string]interface{}) map[string]interface{} {
	for key, value := range m {
		m[key] = ParseLiteral(value)
	}

	return m
}

func walkSlice(items []interface{}) []interface{} {
	for i, item := range items {
		items[i] = ParseLiteral(item)
	}

	return items
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
			return walkMap(value)
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
			var asset skydb.Asset
			mapFromOrPanic((*MapAsset)(&asset), value)
			return &asset
		case "ref":
			var ref skydb.Reference
			mapFromOrPanic((*MapReference)(&ref), value)
			return ref
		case "date":
			var t time.Time
			mapFromOrPanic((*MapTime)(&t), value)
			return t
		case "geo":
			var loc skydb.Location
			mapFromOrPanic((*MapLocation)(&loc), value)
			return loc
		case "geojson":
			var geom skydb.Geometry
			mapFromOrPanic((*MapGeometry)(&geom), value)
			return geom
		case "seq":
			return skydb.Sequence{}
		case "unknown":
			var val skydb.Unknown
			mapFromOrPanic((*MapUnknown)(&val), value)
			return val
		case "relation":
			var rel MapRelation
			mapFromOrPanic((*MapRelation)(&rel), value)
			return &rel
		default:
			panic(fmt.Errorf("unknown $type = %s", kind))
		}
	case []interface{}:
		return walkSlice(value)
	}
}

func mapFromOrPanic(fromMapper FromMapper, m map[string]interface{}) {
	if err := fromMapper.FromMap(m); err != nil {
		panic(err)
	}
}
