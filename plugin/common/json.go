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
	"encoding/json"
	"fmt"

	"github.com/skygeario/skygear-server/skyerr"
)

// ExecError is error resulted from application logic of plugin (e.g.
// an exception thrown within a lambda function)
type ExecError struct {
	// Variable names are prefixed with "Error" because they
	// must be exported for JSON unmarshalling to work, and that
	// they must not be in conflict with the functions of the same name.
	ErrorCode    skyerr.ErrorCode       `json:"code"`
	ErrorMessage string                 `json:"message"`
	ErrorInfo    map[string]interface{} `json:"info"`
}

func (e *ExecError) Name() string {
	return fmt.Sprintf("%v", e.ErrorCode)
}

func (e *ExecError) Code() skyerr.ErrorCode {
	if e.ErrorCode == 0 {
		return skyerr.UnexpectedError
	}
	return e.ErrorCode
}

func (e *ExecError) Message() string {
	if e.ErrorMessage == "" {
		return "An unexpected error has occurred in the plugin."
	}
	return e.ErrorMessage
}

func (e *ExecError) Error() string {
	return fmt.Sprintf("%v: %v", e.Code(), e.Message())
}

func (e *ExecError) Info() map[string]interface{} {
	return e.ErrorInfo
}

func (e *ExecError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name    string                 `json:"name"`
		Code    skyerr.ErrorCode       `json:"code"`
		Message string                 `json:"message"`
		Info    map[string]interface{} `json:"info,omitempty"`
	}{e.Name(), e.Code(), e.Message(), e.Info()})
}
