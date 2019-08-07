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
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
)

// Identity is a principal of user
type Identity struct {
	principal.Attributes
	ID     string
	Type   string
	Claims principal.Claims
}

func NewIdentity(identityProvider principal.IdentityProvider, principal principal.Principal) Identity {
	return Identity{
		ID:         principal.PrincipalID(),
		Type:       principal.ProviderID(),
		Claims:     principal.Claims(),
		Attributes: principal.Attributes(),
	}
}

func (identity Identity) MarshalJSON() ([]byte, error) {
	attrs := map[string]interface{}{}
	for key, value := range identity.Attributes {
		attrs[key] = value
	}
	attrs["id"] = identity.ID
	attrs["type"] = identity.Type
	attrs["claims"] = identity.Claims

	return json.Marshal(attrs)
}

// @JSONSchema
const IdentitySchema = `
{
	"$id": "#Identity",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"type": { "type": "string" },
		"claims": { "type": "object" }
	}
}
`
