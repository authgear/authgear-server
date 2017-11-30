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

package pq

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/paulmach/go.geo"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// This file implements Record related operations of the
// skydb/pq implementation.

// Different data types that can be saved in and loaded from postgreSQL
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.5/static/datatype-character.html
const (
	TypeString                = "text"
	TypeCaseInsensitiveString = "citext"
	TypeNumber                = "double precision"
	TypeBoolean               = "boolean"
	TypeJSON                  = "jsonb"
	TypeTimestamp             = "timestamp without time zone"
	TypeLocation              = "geometry(Point)"
	TypeInteger               = "integer"
	TypeSerial                = "serial UNIQUE"
	TypeBigInteger            = "bigint"
	TypeGeometry              = "geometry"
)

func pqDataType(dataType skydb.DataType) string {
	switch dataType {
	default:
		panic(fmt.Sprintf("Unsupported dataType = %s", dataType))
	case skydb.TypeString, skydb.TypeAsset, skydb.TypeReference:
		return TypeString
	case skydb.TypeNumber:
		return TypeNumber
	case skydb.TypeInteger:
		return TypeInteger
	case skydb.TypeDateTime:
		return TypeTimestamp
	case skydb.TypeBoolean:
		return TypeBoolean
	case skydb.TypeJSON:
		return TypeJSON
	case skydb.TypeLocation:
		return TypeLocation
	case skydb.TypeSequence:
		return TypeSerial
	case skydb.TypeGeometry:
		return TypeGeometry
	}
}

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

	assetName, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan Asset: got type(value) = %T, expect []byte", value)
	}

	na.Asset = &skydb.Asset{
		Name: assetName,
	}
	na.Valid = true

	return nil
}

type nullLocation struct {
	Location skydb.Location
	Valid    bool
}

type nullGeometry struct {
	Geometry skydb.Geometry
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

func (ng *nullGeometry) Scan(value interface{}) error {
	data, ok := value.(string)

	if value == nil || !ok {
		ng.Geometry = nil
		ng.Valid = false
		return nil
	}

	err := json.Unmarshal([]byte(data), &ng.Geometry)
	ng.Valid = err == nil

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

type geometryValue skydb.Geometry

func (geom geometryValue) Value() (driver.Value, error) {
	return json.Marshal(geom)
}

type nullUnknown struct {
	Valid bool
}

func (nu *nullUnknown) Scan(value interface{}) error {
	nu.Valid = value != nil
	return nil
}

type tokenResponseValue struct {
	TokenResponse skydb.TokenResponse
	Valid         bool
}

func (v tokenResponseValue) Value() (driver.Value, error) {
	if !v.Valid {
		return nil, nil
	}

	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(v.TokenResponse); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (v *tokenResponseValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		log.Errorf("skydb: unsupported Scan pair: %T -> %T", value, v.TokenResponse)
	}

	err := json.Unmarshal(b, &v.TokenResponse)
	if err == nil {
		v.Valid = true
	}
	return err
}

type providerProfileValue struct {
	ProviderProfile skydb.ProviderProfile
	Valid           bool
}

func (v providerProfileValue) Value() (driver.Value, error) {
	if !v.Valid {
		return nil, nil
	}

	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(v.ProviderProfile); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (v *providerProfileValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		log.Errorf("skydb: unsupported Scan pair: %T -> %T", value, v.ProviderProfile)
	}

	err := json.Unmarshal(b, &v.ProviderProfile)
	if err == nil {
		v.Valid = true
	}
	return err
}
