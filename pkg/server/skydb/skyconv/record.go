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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// JSONRecord defines a common serialization format for skydb.Record
type JSONRecord skydb.Record

// MarshalJSON implements json.Marshaler
// nolint: gocyclo
func (record *JSONRecord) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	record.ToMap(m)
	return json.Marshal(m)
}

func (record *JSONRecord) ToMap(m map[string]interface{}) {
	for key, value := range record.Data {
		m[key] = ToLiteral(value)
	}

	m["_id"] = record.ID.String() // NOTE(cheungpat): Fields to be deprecated.
	m["_type"] = "record"

	m["_recordID"] = record.ID.Key
	m["_recordType"] = record.ID.Type
	m["_access"] = record.ACL

	if record.OwnerID != "" {
		m["_ownerID"] = record.OwnerID
	}
	if !record.CreatedAt.IsZero() {
		m["_created_at"] = record.CreatedAt
	}
	if record.CreatorID != "" {
		m["_created_by"] = record.CreatorID
	}
	if !record.UpdatedAt.IsZero() {
		m["_updated_at"] = record.UpdatedAt
	}
	if record.UpdaterID != "" {
		m["_updated_by"] = record.UpdaterID
	}

	transient := record.marshalTransient(record.Transient)
	if len(transient) > 0 {
		m["_transient"] = transient
	}
}

func (record *JSONRecord) marshalTransient(transient map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for key, value := range transient {
		m[key] = ToLiteral(value)
	}
	return m
}

// UnmarshalJSON implements json.Unmarshaler
func (record *JSONRecord) UnmarshalJSON(data []byte) (err error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	return record.FromMap(m)
}

func (record *JSONRecord) FromMap(m map[string]interface{}) error {
	var (
		id               skydb.RecordID
		acl              skydb.RecordACL
		ownerID          string
		createdAt        time.Time
		creatorID        string
		updatedAt        time.Time
		updaterID        string
		dataMap          map[string]interface{}
		transientDataMap map[string]interface{}
	)

	extractor := newMapExtractor(m)
	extractor.DoString("_recordID", func(s string) error {
		id.Key = s
		return nil
	}, false)
	extractor.DoString("_recordType", func(s string) error {
		id.Type = s
		return nil
	}, false)
	if id.Key == "" && id.Type == "" {
		// NOTE(cheungpat): Handling for deprecated fields.
		if _, ok := m["_id"]; ok {
			extractor.DoString("_id", func(s string) error {
				return id.UnmarshalText([]byte(s))
			}, true)
			if extractor.Err() != nil {
				return extractor.Err()
			}
		}
	}
	if id.Type == "" {
		return errors.New("missing _recordType, expecting string")
	}
	if id.Key == "" {
		return errors.New("missing _recordID, expecting string")
	}
	extractor.DoString("_ownerID", func(s string) error {
		ownerID = s
		return nil
	}, false)
	extractor.DoTime("_created_at", func(t time.Time) error {
		createdAt = t
		return nil
	}, false)
	extractor.DoString("_created_by", func(s string) error {
		creatorID = s
		return nil
	}, false)
	extractor.DoTime("_updated_at", func(t time.Time) error {
		updatedAt = t
		return nil
	}, false)
	extractor.DoString("_updated_by", func(s string) error {
		updaterID = s
		return nil
	}, false)
	extractor.DoSliceMap("_access", func(slice []map[string]interface{}) error {
		if slice == nil {
			return nil
		}

		acl = skydb.RecordACL{}
		for i, v := range slice {
			ace := skydb.RecordACLEntry{}
			if err := (*MapACLEntry)(&ace).FromMap(v); err != nil {
				return fmt.Errorf(`invalid access entry at %d: %v`, i, err)
			}
			acl = append(acl, ace)
		}
		return nil
	}, false)
	extractor.DoMap("_transient", func(theMap map[string]interface{}) error {
		if theMap == nil {
			return nil
		}
		return (*MapData)(&transientDataMap).FromMap(theMap)
	}, false)
	if extractor.Err() != nil {
		return extractor.Err()
	}

	m = removeReserved(m)
	if err := (*MapData)(&dataMap).FromMap(m); err != nil {
		return err
	}

	record.ID = id
	record.OwnerID = ownerID
	record.CreatedAt = createdAt
	record.CreatorID = creatorID
	record.UpdatedAt = updatedAt
	record.UpdaterID = updaterID
	record.ACL = acl
	record.Transient = transientDataMap
	record.Data = dataMap
	return nil
}

func removeReserved(m map[string]interface{}) map[string]interface{} {
	mm := map[string]interface{}{}
	for key, value := range m {
		if key[0] != '_' {
			mm[key] = value
		}
	}
	return mm
}

// mapExtractor helps to extract value of a key from a map
//
// potential candicate of a package
type mapExtractor struct {
	m   map[string]interface{}
	err error
}

func newMapExtractor(m map[string]interface{}) *mapExtractor {
	return &mapExtractor{m: m}
}

// Do execute doFunc if key exists in the map
// The key will always be removed no matter error occurred previously
func (e *mapExtractor) Do(key string, doFunc func(interface{}) error, required bool) {
	value, ok := e.m[key]
	delete(e.m, key)

	if e.err != nil {
		return
	}

	if ok {
		e.err = doFunc(value)
		delete(e.m, key)
	} else if required {
		e.err = fmt.Errorf(`no key "%s" in map`, key)
	}
}

func (e *mapExtractor) DoString(key string, doFunc func(string) error, required bool) {
	e.Do(key, func(i interface{}) error {
		if m, ok := i.(string); ok {
			return doFunc(m)
		}
		return fmt.Errorf("key %s is of type %T, not string", key, i)
	}, required)
}

func (e *mapExtractor) DoTime(key string, doFunc func(time.Time) error, required bool) {
	e.Do(key, func(i interface{}) error {
		dateStr, ok := i.(string)
		if !ok {
			return fmt.Errorf("key %s is of type %T, not string", key, i)
		}
		dt, err := time.Parse(time.RFC3339Nano, dateStr)
		if err != nil {
			return fmt.Errorf("key %s is not a time: %s", key, err)
		}

		return doFunc(dt)
	}, required)
}

func (e *mapExtractor) DoMap(key string, doFunc func(map[string]interface{}) error, required bool) {
	e.Do(key, func(i interface{}) error {
		if m, ok := i.(map[string]interface{}); ok {
			return doFunc(m)
		}
		return fmt.Errorf("key %s is of type %T, not map[string]interface{}", key, i)
	}, required)
}

func (e *mapExtractor) DoSlice(key string, doFunc func([]interface{}) error, required bool) {
	e.Do(key, func(i interface{}) error {
		switch slice := i.(type) {
		case []interface{}:
			return doFunc(slice)
		case nil:
			return doFunc(nil)
		default:
			return fmt.Errorf("key %s is of type %T, not []interface{}", key, i)
		}
	}, required)
}

func (e *mapExtractor) DoSliceMap(key string, doFunc func([]map[string]interface{}) error, required bool) {
	e.Do(key, func(i interface{}) error {
		switch slice := i.(type) {
		case []interface{}:
			m := []map[string]interface{}{}
			for _, v := range slice {
				if typed, ok := v.(map[string]interface{}); ok {
					m = append(m, typed)
				} else {
					return fmt.Errorf("key %s is of type %T, not []map[string]interface{}", key, i)
				}
			}
			return doFunc(m)
		case nil:
			return doFunc(nil)
		default:
			return fmt.Errorf("key %s is of type %T, not []map[string]interface{}", key, i)
		}
	}, required)
}

func (e *mapExtractor) Err() error {
	return e.err
}
