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
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
)

// Identity is an identity of user
type Identity struct {
	Type   string                 `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

func NewIdentity(info identity.Info) *Identity {
	return &Identity{
		Type:   string(info.Type),
		Claims: info.Claims,
	}
}

// @JSONSchema
const IdentitySchema = `
{
	"$id": "#Identity",
	"type": "object",
	"properties": {
		"type": { "type": "string" },
		"claims": { "type": "object" }
	}
}
`
