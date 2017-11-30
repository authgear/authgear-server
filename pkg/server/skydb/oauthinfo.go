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
	"time"
)

type TokenResponse map[string]interface{}
type ProviderProfile map[string]interface{}

// OAuthInfo contains 3rd provider information for authentication
//
// UserID is AuthInfo ID which incidcate user who link with
// the given oauth data
type OAuthInfo struct {
	UserID          string          `json:"user_id"`
	Provider        string          `json:"provider"`
	PrincipalID     string          `json:"principal_id"`
	TokenResponse   TokenResponse   `json:"token_response,omitempty"`
	ProviderProfile ProviderProfile `json:"profile,omitempty"`
	CreatedAt       *time.Time      `json:"created_at,omitempty"`
	UpdatedAt       *time.Time      `json:"updated_at,omitempty"`
}
