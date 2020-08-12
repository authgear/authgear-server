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
	"database/sql/driver"
	"encoding/json"
)

type NullJSON struct {
	JSON  interface{}
	Valid bool
}

func (nj *NullJSON) Scan(value interface{}) error {
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

type NullJSONMapBoolean struct {
	JSON  map[string]bool
	Valid bool
}

func (nj *NullJSONMapBoolean) Scan(value interface{}) error {
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

// NullJSONStringSlice will reject empty member, since pq will give [null]
// array if we use `array_to_json` on null column. So the result slice will be
// []string{}, but not []string{""}
type NullJSONStringSlice struct {
	Slice []string
	Valid bool
}

func (njss *NullJSONStringSlice) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if value == nil || !ok {
		njss.Slice = nil
		njss.Valid = false
		return nil
	}

	njss.Slice = []string{}
	allSlice := []string{}
	err := json.Unmarshal(data, &allSlice)
	for _, s := range allSlice {
		if s != "" {
			njss.Slice = append(njss.Slice, s)
		}
	}
	njss.Valid = err == nil
	return err
}

type JSONSliceValue []interface{}

func (s JSONSliceValue) Value() (driver.Value, error) {
	return json.Marshal([]interface{}(s))
}

type JSONMapValue map[string]interface{}

func (m JSONMapValue) Value() (driver.Value, error) {
	return json.Marshal(map[string]interface{}(m))
}

type JSONMapBooleanValue map[string]bool

func (m JSONMapBooleanValue) Value() (driver.Value, error) {
	return json.Marshal(map[string]bool(m))
}
