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
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// s3Store implements Store by storing files on S3
type s3Store struct {
	svc       *s3.S3
	uploader  *s3manager.Uploader
	bucket    *string
	urlPrefix string
	public    bool
}

// NewS3Store returns a new s3Store
func NewS3Store(
	accessKey string,
	secretKey string,
	regionName string,
	bucketName string,
	urlPrefix string,
	public bool,
) (Store, error) {

	creds := credentials.NewStaticCredentials(
		accessKey,
		secretKey,
		"",
	)
	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Region:      aws.String(regionName),
		Credentials: creds,
	})
	uploader := s3manager.NewUploaderWithClient(svc)

	bucket := aws.String(bucketName)

	return &s3Store{
		svc:       svc,
		uploader:  uploader,
		bucket:    bucket,
		urlPrefix: urlPrefix,
		public:    public,
	}, nil
}

// GetFileReader returns a reader for files
func (s *s3Store) GetFileReader(name string) (io.ReadCloser, error) {
	key := aws.String(name)
	input := &s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    key,
	}
	objOutput, err := s.svc.GetObject(input)
	return objOutput.Body, err
}

func (s *s3Store) GetRangedFileReader(
	name string,
	fileRange FileRange,
) (*FileRangedGetResult, error) {
	key := aws.String(name)
	rangeHeader := fmt.Sprintf("bytes=%d-%d", fileRange.From, fileRange.To)
	input := &s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    key,
		Range:  &rangeHeader,
	}

	output, err := s.svc.GetObject(input)
	if err != nil {
		return nil, err
	}

	if output.ContentRange == nil {
		return nil, errors.New("missing content ranges header")
	}

	acceptedRange, totalSize, err := parseContentRange(*output.ContentRange)
	if err != nil {
		return nil, err
	}

	return &FileRangedGetResult{
		ReadCloser:    output.Body,
		AcceptedRange: acceptedRange,
		TotalSize:     totalSize,
	}, nil
}

// PutFileReader uploads a file to s3 with content from io.Reader
func (s *s3Store) PutFileReader(
	name string,
	src io.Reader,
	length int64,
	contentType string,
) error {
	key := aws.String(name)
	input := &s3manager.UploadInput{
		Body:        src,
		Bucket:      s.bucket,
		Key:         key,
		ContentType: aws.String(contentType),
	}
	_, err := s.uploader.Upload(input)
	return err
}

// GeneratePostFileRequest return a PostFileRequest for uploading asset
func (s *s3Store) GeneratePostFileRequest(name string, contentType string, length int64) (*PostFileRequest, error) {
	return &PostFileRequest{
		Action: "/files/" + name,
	}, nil
}

// SignedURL return a signed s3 URL with expiry date
func (s *s3Store) SignedURL(name string) (string, error) {
	if !s.IsSignatureRequired() {
		if s.urlPrefix != "" {
			return strings.Join([]string{s.urlPrefix, name}, "/"), nil
		}
		key := aws.String(name)
		input := &s3.GetObjectInput{
			Bucket: s.bucket,
			Key:    key,
		}
		req, _ := s.svc.GetObjectRequest(input)
		// Sign will interpolate the URL String, otherwise the URL will be %bucket%
		req.Sign()
		return req.HTTPRequest.URL.String(), nil
	}
	key := aws.String(name)
	input := &s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    key,
	}
	req, _ := s.svc.GetObjectRequest(input)
	return req.Presign(time.Minute * time.Duration(15))
}

// IsSignatureRequired indicates whether a signature is required
func (s *s3Store) IsSignatureRequired() bool {
	return !s.public
}

// ParseSignature tries to parse the asset signature
func (s *s3Store) ParseSignature(
	signed string,
	name string,
	expiredAt time.Time,
) (bool, error) {

	return false, errors.New(
		"Asset signature parsing for s3-based asset store is not available",
	)
}

// parseContentRange parses the content range string
// in the format of something like `bytes 123-567/1024`
//
func parseContentRange(contentRangeString string) (FileRange, int64, error) {
	splits := strings.SplitN(contentRangeString, " ", 2)
	if len(splits) != 2 {
		return FileRange{}, 0, errors.New("content range is malformed")
	}

	if strings.ToLower(splits[0]) != "bytes" {
		return FileRange{}, 0, errors.New(
			"only support content range in unit of bytes",
		)
	}

	compSplits := strings.SplitN(splits[1], "/", 2)
	if len(compSplits) != 2 {
		return FileRange{}, 0, errors.New("content range is malformed")
	}

	rangeSplits := strings.SplitN(compSplits[0], "-", 2)
	if len(rangeSplits) != 2 {
		return FileRange{}, 0, errors.New("content range is malformed")
	}

	rangeFrom, err1 := strconv.ParseInt(rangeSplits[0], 10, 64)
	rangeTo, err2 := strconv.ParseInt(rangeSplits[1], 10, 64)
	totalSize, err3 := strconv.ParseInt(compSplits[1], 10, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return FileRange{}, 0, errors.New("content range is malformed")
	}

	return FileRange{
		From: rangeFrom,
		To:   rangeTo,
	}, totalSize, nil
}
