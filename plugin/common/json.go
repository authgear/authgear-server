package common

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oursky/skygear/oddb"
	"github.com/oursky/skygear/oddb/oddbconv"
)

// ExecError is error resulted from application logic of plugin (e.g.
// an exception thrown within a lambda function)
type ExecError struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

func (err *ExecError) Error() string {
	return err.Name + "\n" + err.Description
}

// JSONRecord defines a common serialization format for oddb.Record
type JSONRecord oddb.Record

// MarshalJSON implements json.Marshaler
func (record *JSONRecord) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	for key, value := range record.Data {
		switch v := value.(type) {
		case time.Time:
			data[key] = (oddbconv.MapTime)(v)
		case oddb.Asset:
			data[key] = (oddbconv.MapAsset)(v)
		case oddb.Reference:
			data[key] = (oddbconv.MapReference)(v)
		case *oddb.Location:
			data[key] = (*oddbconv.MapLocation)(v)
		default:
			data[key] = value
		}
	}

	m := map[string]interface{}{}
	oddbconv.MapData(data).ToMap(m)

	m["_id"] = record.ID
	m["_ownerID"] = record.OwnerID
	m["_access"] = record.ACL

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

	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler
func (record *JSONRecord) UnmarshalJSON(data []byte) (err error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	var (
		id      oddb.RecordID
		acl     oddb.RecordACL
		dataMap map[string]interface{}
	)

	extractor := newMapExtractor(m)
	extractor.DoString("_id", func(s string) error {
		return id.UnmarshalText([]byte(s))
	})
	extractor.DoSlice("_access", func(slice []interface{}) error {
		return acl.InitFromJSON(slice)
	})
	if extractor.Err() != nil {
		return extractor.Err()
	}

	m = sanitizedDataMap(m)
	if err := (*oddbconv.MapData)(&dataMap).FromMap(m); err != nil {
		return err
	}

	record.ID = id
	record.ACL = acl
	record.Data = dataMap
	return nil
}

func sanitizedDataMap(m map[string]interface{}) map[string]interface{} {
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
func (e *mapExtractor) Do(key string, doFunc func(interface{}) error) {
	value, ok := e.m[key]
	delete(e.m, key)

	if e.err != nil {
		return
	}

	if ok {
		e.err = doFunc(value)
		delete(e.m, key)
	} else {
		e.err = fmt.Errorf(`no key "%s" in map`, key)
	}
}

func (e *mapExtractor) DoString(key string, doFunc func(string) error) {
	e.Do(key, func(i interface{}) error {
		if m, ok := i.(string); ok {
			return doFunc(m)
		}
		return fmt.Errorf("key %s is of type %T, not string", key, i)
	})
}

func (e *mapExtractor) DoMap(key string, doFunc func(map[string]interface{}) error) {
	e.Do(key, func(i interface{}) error {
		if m, ok := i.(map[string]interface{}); ok {
			return doFunc(m)
		}
		return fmt.Errorf("key %s is of type %T, not map[string]interface{}", key, i)
	})
}

func (e *mapExtractor) DoSlice(key string, doFunc func([]interface{}) error) {
	e.Do(key, func(i interface{}) error {
		switch slice := i.(type) {
		case []interface{}:
			return doFunc(slice)
		case nil:
			return doFunc(nil)
		default:
			return fmt.Errorf("key %s is of type %T, not []interface{}", key, i)
		}
	})
}

func (e *mapExtractor) Err() error {
	return e.err
}
