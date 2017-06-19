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
	"fmt"
	"strings"
)

// TraverseColumnTypes traverse the field type of a key path from database table.
func TraverseColumnTypes(db Database, recordType string, keyPath string) ([]FieldType, error) {
	fields := []FieldType{}
	components := strings.Split(keyPath, ".")
	for i, component := range components {
		field := FieldType{}
		isLast := (i == len(components)-1)

		schema, err := db.RemoteColumnTypes(recordType)
		if err != nil {
			return fields, fmt.Errorf(`record type "%s" does not exist`, recordType)
		}

		if f, ok := schema[component]; ok {
			field = f
		} else {
			return fields, fmt.Errorf(`keypath "%s" does not exist`, keyPath)
		}

		if field.Type != TypeReference && !isLast {
			return fields, fmt.Errorf(`field "%s" in keypath "%s" is not a reference`, component, keyPath)
		}

		fields = append(fields, field)

		if field.Type == TypeReference {
			recordType = field.ReferenceType
		}
	}
	return fields, nil
}
