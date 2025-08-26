package cloudstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

type AlibabaCloudOSSStorage struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string

	cfg    *oss.Config
	client *oss.Client
}

var _ storage = &AlibabaCloudOSSStorage{}

func NewAlibabaCloudOSSStorage(accessKeyID, secretAccessKey, region, bucket string) (*AlibabaCloudOSSStorage, error) {
	cred := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey)
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(cred).
		WithRegion(region)

	client := oss.NewClient(cfg)

	return &AlibabaCloudOSSStorage{
		Region:          region,
		Bucket:          bucket,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,

		cfg:    cfg,
		client: client,
	}, nil
}

func (s *AlibabaCloudOSSStorage) PresignPutObject(ctx context.Context, name string, header http.Header) (*http.Request, error) {
	input := &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.Bucket),
		Key:    oss.Ptr(name),
	}

	for name := range header {
		lower := strings.ToLower(name)
		switch lower {
		case "content-type":
			input.ContentType = oss.Ptr(header.Get(name))
		case "content-disposition":
			input.ContentDisposition = oss.Ptr(header.Get(name))
		case "content-encoding":
			input.ContentEncoding = oss.Ptr(header.Get(name))
		case "content-length":
			contentLengthStr := header.Get(name)
			contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content-length: %w", err)
			}
			input.ContentLength = oss.Ptr(contentLength)
		case "content-md5":
			input.ContentMD5 = oss.Ptr(header.Get(name))
		case "cache-control":
			input.CacheControl = oss.Ptr(header.Get(name))
		}
	}

	result, err := s.client.Presign(ctx, input, oss.PresignExpires(PresignPutExpires))
	if err != nil {
		return nil, fmt.Errorf("failed to presign put object: %w", err)
	}

	u, err := url.Parse(result.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	return &http.Request{
		Method: result.Method,
		URL:    u,
		// Actually, result.SignedHeaders is a map[string]string that we could possibly return.
		// From my observation, the difference between header and result.SignedHeaders is that
		// result.SignedHeaders does not contain "Content-Length".
		// So we return header here.
		Header: header,
	}, nil
}

func (s *AlibabaCloudOSSStorage) PresignHeadObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	input := &oss.HeadObjectRequest{
		Bucket: oss.Ptr(s.Bucket),
		Key:    oss.Ptr(name),
	}

	result, err := s.client.Presign(ctx, input, oss.PresignExpires(expire))
	if err != nil {
		return nil, fmt.Errorf("failed to presign head request: %w", err)
	}

	u, err := url.Parse(result.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	return u, nil
}

func (s *AlibabaCloudOSSStorage) PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	input := &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.Bucket),
		Key:    oss.Ptr(name),
	}

	result, err := s.client.Presign(ctx, input, oss.PresignExpires(expire))
	if err != nil {
		return nil, fmt.Errorf("failed to presign get request: %w", err)
	}

	u, err := url.Parse(result.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	return u, nil
}

func (s *AlibabaCloudOSSStorage) MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request) {
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
