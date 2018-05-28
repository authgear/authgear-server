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

package asset

import (
	"io"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/logging"
)

var log = logging.LoggerEntry("asset")

// PostFileRequest models the POST request for upload asset file
type PostFileRequest struct {
	Action      string                 `json:"action"`
	ExtraFields map[string]interface{} `json:"extra-fields,omitempty"`
}

// Store specify the interfaces of an asset store
type Store interface {
	GetFileReader(name string) (io.ReadCloser, error)
	PutFileReader(name string, src io.Reader, length int64, contentType string) error
	GeneratePostFileRequest(name string, contentType string, length int64) (*PostFileRequest, error)
}

// URLSigner signs a signature and returns a URL accessible to that asset.
type URLSigner interface {
	// SignedURL returns a url with access to the named file. If asset
	// store is private, the returned URL is a signed one, allowing access
	// to asset for a short period.
	SignedURL(name string) (string, error)
	IsSignatureRequired() bool
}

// URLSignerStore is an interface that is a union of Store and URLSigner.
//go:generate mockgen -destination=mock_asset/mock_url_signer_store.go github.com/skygeario/skygear-server/pkg/server/asset URLSignerStore
type URLSignerStore interface {
	Store
	URLSigner
}

// SignatureParser parses a signed signature string
type SignatureParser interface {
	ParseSignature(signed string, name string, expiredAt time.Time) (valid bool, err error)
}
