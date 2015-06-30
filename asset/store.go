package asset

import (
	"io"
	"time"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

// Store specify the interfaces of an asset store
type Store interface {
	PutFileReader(name string, src io.Reader, length int64, contentType string) error

	// SignedURL returns a signed url with access to the named file. The link
	// should expires itself after expiredAt
	SignedURL(name string, expiredAt time.Time) string
}

// S3Store implements Store by storing files on S3
type S3Store struct {
	bucket *s3.Bucket
}

// NewS3Store returns a new S3Store
func NewS3Store(accessKey, secretKey, bucket string) *S3Store {
	auth := aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
	return &S3Store{
		// FIXME(limouren): auto detect aws region
		bucket: s3.New(auth, aws.APNortheast).Bucket(bucket),
	}
}

// PutFileReader uploads a file to s3 with content from io.Reader
func (s *S3Store) PutFileReader(name string, src io.Reader, length int64, contentType string) error {
	return s.bucket.PutReader(name, src, length, contentType, s3.Private)
}

// SignedURL return a signed s3 URL with expiry date
func (s *S3Store) SignedURL(name string, expiredAt time.Time) string {
	return s.bucket.SignedURL(name, expiredAt)
}
