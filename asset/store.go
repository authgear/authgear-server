package asset

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

// Store specify the interfaces of an asset store
type Store interface {
	GetFileReader(name string) (io.ReadCloser, error)
	PutFileReader(name string, src io.Reader, length int64, contentType string) error
}

type URLSigner interface {
	// SignedURL returns a signed url with access to the named file. The link
	// should expires itself after expiredAt
	SignedURL(name string, expiredAt time.Time) string
}

// FileStore implements Store by storing files on file system
type FileStore struct {
	dir string
}

func NewFileStore(dir string) *FileStore {
	return &FileStore{dir}
}

func (s *FileStore) GetFileReader(name string) (io.ReadCloser, error) {
	path := filepath.Join(s.dir, name)
	return os.Open(path)
}

// PutFileReader stores a file from reader onto file system
func (s *FileStore) PutFileReader(name string, src io.Reader, length int64, contentType string) error {
	path := filepath.Join(s.dir, name)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	written, err := io.Copy(f, src)
	if err != nil {
		return err
	}

	if written != length {
		return fmt.Errorf("got written %d bytes, expect %d", written, length)
	}

	return nil
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

func (s *S3Store) GetFileReader(name string) (io.ReadCloser, error) {
	return s.bucket.GetReader(name)
}

// PutFileReader uploads a file to s3 with content from io.Reader
func (s *S3Store) PutFileReader(name string, src io.Reader, length int64, contentType string) error {
	return s.bucket.PutReader(name, src, length, contentType, s3.Private)
}

// SignedURL return a signed s3 URL with expiry date
func (s *S3Store) SignedURL(name string, expiredAt time.Time) string {
	return s.bucket.SignedURL(name, expiredAt)
}
