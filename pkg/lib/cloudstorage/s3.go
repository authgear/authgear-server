package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Storage struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string

	session *session.Session
	s3      *s3.S3
}

var _ Storage = &S3Storage{}

func NewS3Storage(accessKeyID, secretAccessKey, region, bucket string) (*S3Storage, error) {
	cred := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	s3 := s3.New(sess)
	return &S3Storage{
		Region:          region,
		Bucket:          bucket,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,

		session: sess,
		s3:      s3,
	}, nil
}

func (s *S3Storage) PresignPutObject(name string, header http.Header) (*http.Request, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}

	metadata := map[string]*string{}
	for name := range header {
		lower := strings.ToLower(name)
		switch lower {
		case "content-type":
			input.SetContentType(header.Get(name))
		case "content-disposition":
			input.SetContentDisposition(header.Get(name))
		case "content-encoding":
			input.SetContentEncoding(header.Get(name))
		case "content-length":
			contentLengthStr := header.Get(name)
			contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content-length: %w", err)
			}
			input.SetContentLength(contentLength)
		case "content-md5":
			input.SetContentMD5(header.Get(name))
		case "cache-control":
			input.SetCacheControl(header.Get(name))
		}
	}
	input.SetMetadata(metadata)

	req, _ := s.s3.PutObjectRequest(input)
	req.NotHoist = true
	urlStr, _, err := req.PresignRequest(PresignPutExpires)
	if err != nil {
		return nil, fmt.Errorf("failed to presign put request: %w", err)
	}
	u, _ := url.Parse(urlStr)

	return &http.Request{
		Method: "PUT",
		URL:    u,
		Header: header,
	}, nil
}

func (s *S3Storage) PresignHeadObject(name string, expire time.Duration) (*url.URL, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}
	req, _ := s.s3.HeadObjectRequest(input)
	req.NotHoist = false
	urlStr, _, err := req.PresignRequest(expire)
	if err != nil {
		return nil, fmt.Errorf("failed to presign head request: %w", err)
	}
	u, _ := url.Parse(urlStr)

	return u, nil
}

func (s *S3Storage) PresignGetObject(name string, expire time.Duration) (*url.URL, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}
	req, _ := s.s3.GetObjectRequest(input)
	req.NotHoist = false
	urlStr, _, err := req.PresignRequest(expire)
	if err != nil {
		return nil, fmt.Errorf("failed to presign get request: %w", err)
	}
	u, _ := url.Parse(urlStr)

	return u, nil
}

func (s *S3Storage) MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)
		input := &s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(key),
		}
		req, _ := s.s3.GetObjectRequest(input)
		req.NotHoist = false
		urlStr, _, err := req.PresignRequest(PresignGetExpires)
		if err != nil {
			panic(fmt.Errorf("failed to presign head request: %w", err))
		}

		u, _ := url.Parse(urlStr)
		r.Host = ""
		r.URL = u
	}
}
