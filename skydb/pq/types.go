package pq

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/oursky/skygear/skydb"
	"github.com/paulmach/go.geo"
)

// This file implements Record related operations of the
// skydb/pq implementation.

// Different data types that can be saved in and loaded from postgreSQL
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.4/static/datatype-character.html
const (
	TypeString    = "text"
	TypeNumber    = "double precision"
	TypeBoolean   = "boolean"
	TypeJSON      = "jsonb"
	TypeTimestamp = "timestamp without time zone"
	TypeLocation  = "geometry(Point)"
	TypeInteger   = "integer"
	TypeSerial    = "serial UNIQUE"
)

type nullJSON struct {
	JSON  interface{}
	Valid bool
}

func (nj *nullJSON) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if value == nil || !ok {
		nj.JSON = nil
		nj.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &nj.JSON)
	nj.Valid = err == nil

	return err
}

// nullJSONStringSlice will reject empty member, since pq will give [null]
// array if we use `array_to_json` on null column. So the result slice will be
// []string{}, but not []string{""}
type nullJSONStringSlice struct {
	slice []string
	Valid bool
}

func (njss *nullJSONStringSlice) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if value == nil || !ok {
		njss.slice = nil
		njss.Valid = false
		return nil
	}

	njss.slice = []string{}
	allSlice := []string{}
	err := json.Unmarshal(data, &allSlice)
	for _, s := range allSlice {
		if s != "" {
			njss.slice = append(njss.slice, s)
		}
	}
	njss.Valid = err == nil
	return err
}

type assetValue skydb.Asset

func (asset assetValue) Value() (driver.Value, error) {
	return asset.Name, nil
}

type nullAsset struct {
	Asset *skydb.Asset
	Valid bool
}

func (na *nullAsset) Scan(value interface{}) error {
	if value == nil {
		na.Asset = &skydb.Asset{}
		na.Valid = false
		return nil
	}

	assetName, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Asset: got type(value) = %T, expect []byte", value)
	}

	na.Asset = &skydb.Asset{
		Name: string(assetName),
	}
	na.Valid = true

	return nil
}

type nullLocation struct {
	Location skydb.Location
	Valid    bool
}

func (nl *nullLocation) Scan(value interface{}) error {
	if value == nil {
		nl.Location = skydb.Location{}
		nl.Valid = false
		return nil
	}

	src, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Location: got type(value) = %T, expect []byte", value)
	}

	// TODO(limouren): instead of decoding a str-encoded hex, we should utilize
	// ST_AsBinary to perform the SELECT
	decoded := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(decoded, src)
	if err != nil {
		return fmt.Errorf("failed to scan Location: malformed wkb")
	}

	err = (*geo.Point)(&nl.Location).Scan(decoded)
	nl.Valid = err == nil
	return err
}

type referenceValue skydb.Reference

func (ref referenceValue) Value() (driver.Value, error) {
	return ref.ID.Key, nil
}

type jsonSliceValue []interface{}

func (s jsonSliceValue) Value() (driver.Value, error) {
	return json.Marshal([]interface{}(s))
}

type jsonMapValue map[string]interface{}

func (m jsonMapValue) Value() (driver.Value, error) {
	return json.Marshal(map[string]interface{}(m))
}

type aclValue skydb.RecordACL

func (acl aclValue) Value() (driver.Value, error) {
	if acl == nil {
		return nil, nil
	}
	return json.Marshal(acl)
}

type locationValue skydb.Location

func (loc locationValue) Value() (driver.Value, error) {
	return geo.Point(loc).ToWKT(), nil
}
