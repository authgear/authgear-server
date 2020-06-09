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

package model

import (
	"time"
)

// User is the unify way of returning a AuthInfo with LoginID to SDK
type User struct {
	ID          string                 `json:"id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	LastLoginAt *time.Time             `json:"last_login_at,omitempty"`
	IsAnonymous bool                   `json:"is_anonymous"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// @JSONSchema
const UserSchema = `
{
	"$id": "#User",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"created_at": { "type": "string" },
		"last_login_at": { "type": "string" },
		"is_anonymous": { "type": "boolean" },
		"metadata": { "type": "object" }
	}
}
`
