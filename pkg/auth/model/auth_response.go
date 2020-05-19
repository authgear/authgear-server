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

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
}

func NewAuthResponse(user User) AuthResponse {
	return AuthResponse{
		User: user,
	}
}

// @JSONSchema
const UserResponseSchema = `
{
	"$id": "#UserResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"user": { "$ref": "#User" }
			}
		}
	}
}
`

// @JSONSchema
const AuthResponseSchema = `
{
	"$id": "#AuthResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"user": { "$ref": "#User" },
				"access_token": { "type": "string" },
				"refresh_token": { "type": "string" },
				"session_id": { "type": "string" }
			}
		}
	}
}
`
