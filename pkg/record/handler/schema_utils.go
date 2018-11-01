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

package handler

import (
	"sort"
	"strings"

	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

type schemaFieldList struct {
	Fields []schemaField `mapstructure:"fields" json:"fields"`
}

func (s schemaFieldList) Len() int {
	return len(s.Fields)
}

func (s schemaFieldList) Swap(i, j int) {
	s.Fields[i], s.Fields[j] = s.Fields[j], s.Fields[i]
}

func (s schemaFieldList) Less(i, j int) bool {
	return strings.Compare(s.Fields[i].Name, s.Fields[j].Name) < 0
}

type schemaField struct {
	Name     string `mapstructure:"name" json:"name"`
	TypeName string `mapstructure:"type" json:"type"`
}

func encodeRecordSchemas(data map[string]record.Schema) map[string]schemaFieldList {
	schemaMap := make(map[string]schemaFieldList)
	for recordType, schema := range data {
		fieldList := schemaFieldList{
			// initialize array so this will marshal as `[]` instead of
			// `null`
			Fields: []schemaField{},
		}
		for fieldName, val := range schema {
			if strings.HasPrefix(fieldName, "_") {
				continue
			}

			fieldList.Fields = append(fieldList.Fields, schemaField{
				Name:     fieldName,
				TypeName: val.ToSimpleName(),
			})
		}
		sort.Sort(fieldList)
		schemaMap[recordType] = fieldList
	}

	return schemaMap
}
