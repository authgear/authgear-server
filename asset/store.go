package asset

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

// Store specify the interfaces of an asset store
type Store interface {
	GetFileReader(name string) (io.ReadCloser, error)
	PutFileReader(name string, src io.Reader, length int64, contentType string) error
}

// URLSigner signs a signature and returns a URL accessible to that asset.
type URLSigner interface {
	// SignedURL returns a signed url with access to the named file. The link
	// should expires itself after expiredAt
	SignedURL(name string, expiredAt time.Time) string
}

// SignatureParser parses a signed signature string
type SignatureParser interface {
	ParseSignature(signed string, name string, expiredAt time.Time) (valid bool, err error)
}

// FileStore implements Store by storing files on file system
type FileStore struct {
	dir    string
	prefix string
	secret string
}

func NewFileStore(dir, prefix, secret string) *FileStore {
	return &FileStore{dir, prefix, secret}
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

func (s *FileStore) SignedURL(name string, expiredAt time.Time) string {
	expiredAtStr := strconv.FormatInt(expiredAt.Unix(), 10)

	h := hmac.New(sha256.New, []byte(s.secret))
	io.WriteString(h, name)
	io.WriteString(h, expiredAtStr)

	buf := bytes.Buffer{}
	base64Encoder := base64.NewEncoder(base64.URLEncoding, &buf)
	base64Encoder.Write(h.Sum(nil))

	return fmt.Sprintf(
		"%s/%s?expiredAt=%s&signature=%s",
		s.prefix, name, expiredAtStr, buf.String(),
	)
}

func (s *FileStore) ParseSignature(signed string, name string, expiredAt time.Time) (valid bool, err error) {
	base64Decoder := base64.NewDecoder(base64.URLEncoding, strings.NewReader(signed))
	remoteSignature, err := ioutil.ReadAll(base64Decoder)
	if err != nil {
		log.Errorf("failed to decode asset url signature: %v", err)

		return false, errors.New("invalid signature")
	}

	h := hmac.New(sha256.New, []byte(s.secret))
	io.WriteString(h, name)
	io.WriteString(h, strconv.FormatInt(expiredAt.Unix(), 10))

	return !hmac.Equal(remoteSignature, h.Sum(nil)), nil
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
