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
	"errors"
	"fmt"
	"io"
	"time"

	"gopkg.in/amz.v3/aws"
	"gopkg.in/amz.v3/s3"
)

// S3Store implements Store by storing files on S3
type S3Store struct {
	bucket *s3.Bucket
	public bool
}

// NewS3Store returns a new S3Store
func NewS3Store(accessKey, secretKey, regionName, bucketName string, public bool) (*S3Store, error) {
	auth := aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	region, ok := aws.Regions[regionName]
	if !ok {
		return nil, fmt.Errorf("unrecgonized region name = %v", regionName)
	}

	bucket, err := s3.New(auth, region).Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return &S3Store{
		bucket: bucket,
		public: public,
	}, nil
}

// GetFileReader returns a reader for files
func (s *S3Store) GetFileReader(name string) (io.ReadCloser, error) {
	return s.bucket.GetReader(name)
}

// PutFileReader uploads a file to s3 with content from io.Reader
func (s *S3Store) PutFileReader(
	name string,
	src io.Reader,
	length int64,
	contentType string,
) error {

	return s.bucket.PutReader(name, src, length, contentType, s3.Private)
}

// SignedURL return a signed s3 URL with expiry date
func (s *S3Store) SignedURL(name string) (string, error) {
	if !s.IsSignatureRequired() {
		return s.bucket.URL(name), nil
	}
	return s.bucket.SignedURL(name, time.Minute*time.Duration(15))
}

// IsSignatureRequired indicates whether a signature is required
func (s *S3Store) IsSignatureRequired() bool {
	return !s.public
}

// ParseSignature tries to parse the asset signature
func (s *S3Store) ParseSignature(
	signed string,
	name string,
	expiredAt time.Time,
) (bool, error) {

	return false, errors.New(
		"Asset signature parsing for s3-based asset store is not available",
	)
}
