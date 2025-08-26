package cloudstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string

	cfg           aws.Config
	s3            *s3.Client
	presignClient *s3.PresignClient
}

var _ storage = &S3Storage{}

func NewS3Storage(accessKeyID, secretAccessKey, region, bucket string) (*S3Storage, error) {
	cred := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(cred),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	return &S3Storage{
		Region:          region,
		Bucket:          bucket,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,

		cfg:           cfg,
		s3:            s3Client,
		presignClient: presignClient,
	}, nil
}

func (s *S3Storage) PresignPutObject(ctx context.Context, name string, header http.Header) (*http.Request, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}

	for name := range header {
		lower := strings.ToLower(name)
		switch lower {
		case "content-type":
			input.ContentType = aws.String(header.Get(name))
		case "content-disposition":
			input.ContentDisposition = aws.String(header.Get(name))
		case "content-encoding":
			input.ContentEncoding = aws.String(header.Get(name))
		case "content-length":
			contentLengthStr := header.Get(name)
			contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content-length: %w", err)
			}
			input.ContentLength = aws.Int64(contentLength)
		case "content-md5":
			input.ContentMD5 = aws.String(header.Get(name))
		case "cache-control":
			input.CacheControl = aws.String(header.Get(name))
		}
	}

	req, err := s.presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(PresignPutExpires))
	if err != nil {
		return nil, fmt.Errorf("failed to presign put object: %w", err)
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	return &http.Request{
		Method: req.Method,
		URL:    u,
		// Actually, req.SignedHeader is also a http.Header that we could possibly return.
		// From my observation, the different between header and req.SignedHeader is that
		// req.SignedHeader contains "Host".
		// So the difference is insignificant because URL contains the host as well.
		Header: header,
	}, nil
}

func (s *S3Storage) PresignHeadObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}

	req, err := s.presignClient.PresignHeadObject(ctx, input, s3.WithPresignExpires(expire))
	if err != nil {
		return nil, fmt.Errorf("failed to presign head object: %w", err)
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	return u, nil
}

func (s *S3Storage) PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}

	req, err := s.presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expire))
	if err != nil {
		return nil, fmt.Errorf("failed to presign get request: %w", err)
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	return u, nil
}

func (s *S3Storage) MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)

		u, err := s.PresignGetObject(r.Context(), key, expire)
		if err != nil {
			panic(fmt.Errorf("failed to presign get object: %w", err))
		}

		r.Host = ""
		r.URL = u
	}
}
