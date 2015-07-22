package oddbconv

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/oursky/ourd/oddb"
)

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

// MapAsset is oddb.Asset that can be converted from and to a map.
type MapAsset oddb.Asset

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
	return nil
}

// MapReference is oddb.Reference that can be converted from and to a map.
type MapReference oddb.Reference

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

func walkMap(m map[string]interface{}) map[string]interface{} {
	for key, value := range m {
		m[key] = ParseInterface(value)
	}

	return m
}

func walkSlice(items []interface{}) []interface{} {
	for i, item := range items {
		items[i] = ParseInterface(item)
	}

	return items
}

// ParseInterface deduces whether i is a oddb data value and returns a
// parsed value.
//
// FIXME(limouren): this function is public because RecordQueryHandler
// needs it to parse predicate's expression
func ParseInterface(i interface{}) interface{} {
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
			panic(fmt.Errorf("unsupported $type of persistence = %s", kind))
		case "geo", "blob":
			panic(fmt.Errorf("unimplemented $type = %s", kind))
		case "asset":
			var asset oddb.Asset
			mapFromOrPanic((*MapAsset)(&asset), value)
			return asset
		case "ref":
			var ref oddb.Reference
			mapFromOrPanic((*MapReference)(&ref), value)
			return ref
		case "date":
			var t time.Time
			mapFromOrPanic((*MapTime)(&t), value)
			return t
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
