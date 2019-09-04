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
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type AuthResponse struct {
	User         User      `json:"user"`
	Identity     *Identity `json:"identity,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	SessionID    string    `json:"session_id,omitempty"`
}

func NewAuthResponseWithUser(user User) AuthResponse {
	return AuthResponse{
		User: user,
	}
}

func NewAuthResponseWithUserIdentity(user User, identity Identity) AuthResponse {
	return AuthResponse{
		User:     user,
		Identity: &identity,
	}
}

func NewAuthResponse(user User, identity Identity, session *auth.Session) AuthResponse {
	return AuthResponse{
		User:         user,
		Identity:     &identity,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		SessionID:    session.ID,
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
const UserIdentityResponseSchema = `
{
	"$id": "#UserIdentityResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"user": { "$ref": "#User" },
				"identity": { "$ref": "#Identity" }
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
				"identity": { "$ref": "#Identity" },
				"access_token": { "type": "string" },
				"refresh_token": { "type": "string" },
				"session_id": { "type": "string" }
			}
		}
	}
}
`
