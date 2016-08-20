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

package common

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"
)

// EncodeBase64JSON encodes an object into a base64 encoded JSON string
func EncodeBase64JSON(data interface{}) (string, error) {
	out, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	result := base64.StdEncoding.EncodeToString(out)
	return result, nil
}

// DecodeBase64JSON decodes a base64 encoded JSON string into an object
func DecodeBase64JSON(encodedStr string, obj interface{}) error {
	decoder := base64.NewDecoder(base64.URLEncoding, strings.NewReader(encodedStr))
	out, err := ioutil.ReadAll(decoder)
	if err != nil {
		return err
	}

	return json.Unmarshal(out, obj)
}
