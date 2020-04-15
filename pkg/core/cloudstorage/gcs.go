package cloudstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

var ErrInvalidCredentialsJSON = errors.New("invalid credentials JSON")

type GCSStorage struct {
	ServiceAccount  string
	Bucket          string
	CredentialsJSON []byte

	privateKey []byte
	client     *storage.Client
	service    *raw.Service
	err        error
}

var _ Storage = &GCSStorage{}

func NewGCSStorage(credentialsJSON []byte, serviceAccount string, bucket string) *GCSStorage {
	s := &GCSStorage{
		ServiceAccount:  serviceAccount,
		Bucket:          bucket,
		CredentialsJSON: credentialsJSON,
	}

	var j map[string]interface{}
	err := json.NewDecoder(bytes.NewReader(credentialsJSON)).Decode(&j)
	if err != nil {
		s.err = errors.HandledWithMessage(err, "failed to parse credentials JSON")
		return s
	}

	privateKeyStr, ok := j["private_key"].(string)
	if !ok {
		err = ErrInvalidCredentialsJSON
		s.err = err
		return s
	}
	s.privateKey = []byte(privateKeyStr)

	ctx := context.Background()
	service, err := raw.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		s.err = errors.HandledWithMessage(err, "failed to initialize GCS")
		return s
	}
	s.service = service

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		s.err = errors.HandledWithMessage(err, "failed to initialize GCS")
		return s
	}
	s.client = client

	return s
}

const (
	GCSHeaderAccess = "x-goog-meta-access"
)

var GCSProprietaryToStandardMap = map[string]string{
	"x-goog-meta-accesscontrolalloworigin":      "access-control-allow-origin",
	"x-goog-meta-accesscontrolexposeheaders":    "access-control-expose-headers",
	"x-goog-meta-accesscontrolmaxage":           "access-control-max-age",
	"x-goog-meta-accesscontrolallowcredentials": "access-control-allow-credentials",
	"x-goog-meta-accesscontrolallowmethods":     "access-control-allow-methods",
	"x-goog-meta-accesscontrolallowheaders":     "access-control-allow-headers",
}

var GCSStandardToProprietaryMap = map[string]string{
	"access-control-allow-origin":      "x-goog-meta-accesscontrolalloworigin",
	"access-control-expose-headers":    "x-goog-meta-accesscontrolexposeheaders",
	"access-control-max-age":           "x-goog-meta-accesscontrolmaxage",
	"access-control-allow-credentials": "x-goog-meta-accesscontrolallowcredentials",
	"access-control-allow-methods":     "x-goog-meta-accesscontrolallowmethods",
	"access-control-allow-headers":     "x-goog-meta-accesscontrolallowheaders",
}

func (s *GCSStorage) PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error) {
	if s.err != nil {
		return nil, s.err
	}

	now := time.Now().UTC()

	header = s.StandardToProprietary(header)
	header.Set(GCSHeaderAccess, string(accessType))

	// We must omit Content-type and Content-MD5 from header because they are special.
	var headers []string
	for name := range header {
		lower := strings.ToLower(name)
		if lower == "content-type" || lower == "content-md5" {
			continue
		}
		headers = append(headers, fmt.Sprintf("%s:%s", lower, header.Get(name)))
	}

	expires := now.Add(PresignPutExpires)
	opts := storage.SignedURLOptions{
		GoogleAccessID: s.ServiceAccount,
		PrivateKey:     s.privateKey,
		Method:         "PUT",
		Expires:        expires,
		ContentType:    header.Get("Content-Type"),
		Headers:        headers,
		MD5:            header.Get("Content-MD5"),
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := storage.SignedURL(s.Bucket, name, &opts)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to presign put request")
	}

	u, _ := url.Parse(urlStr)
	req := http.Request{
		Method: "PUT",
		Header: header,
		URL:    u,
	}

	return &req, nil
}

func (s *GCSStorage) PresignGetOrHeadObject(name string, method string) (*url.URL, error) {
	if s.err != nil {
		return nil, s.err
	}

	now := time.Now().UTC()
	expires := now.Add(PresignGetExpires)

	opts := storage.SignedURLOptions{
		GoogleAccessID: s.ServiceAccount,
		PrivateKey:     s.privateKey,
		Method:         method,
		Expires:        expires,
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := storage.SignedURL(s.Bucket, name, &opts)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to presign get or head request")
	}

	u, _ := url.Parse(urlStr)

	return u, nil
}

func (s *GCSStorage) PresignGetObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "GET")
}

func (s *GCSStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "HEAD")
}

func (s GCSStorage) ListObjects(r *ListObjectsRequest) (*ListObjectsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	call := s.service.Objects.List(s.Bucket)
	call.Projection("full")
	call.MaxResults(int64(r.PageSize))
	call.Prefix(r.Prefix)
	if r.PaginationToken != "" {
		call.PageToken(r.PaginationToken)
	}

	objects, err := call.Do()
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list objects")
	}

	resp := &ListObjectsResponse{}

	resp.PaginationToken = objects.NextPageToken

	for _, item := range objects.Items {
		resp.Assets = append(resp.Assets, ListAssetItem{
			AssetName: item.Name,
			Size:      int64(item.Size),
		})
	}
	if resp.Assets == nil {
		resp.Assets = []ListAssetItem{}
	}

	return resp, nil
}

func (s *GCSStorage) DeleteObject(name string) error {
	if s.err != nil {
		return s.err
	}
	bucketHandle := s.client.Bucket(s.Bucket)
	objectHandle := bucketHandle.Object(name)
	ctx := context.Background()
	err := objectHandle.Delete(ctx)
	if err == storage.ErrObjectNotExist {
		return nil
	}
	if err != nil {
		return errors.HandledWithMessage(err, "failed to delete object")
	}
	return nil
}

func (s *GCSStorage) StandardToProprietary(header http.Header) http.Header {
	return RewriteHeaderName(header, GCSStandardToProprietaryMap)
}

func (s *GCSStorage) ProprietaryToStandard(header http.Header) http.Header {
	return RewriteHeaderName(header, GCSProprietaryToStandardMap)
}

func (s *GCSStorage) AccessType(header http.Header) AccessType {
	a := header.Get(GCSHeaderAccess)
	switch a {
	case string(AccessTypePublic):
		return AccessTypePublic
	case string(AccessTypePrivate):
		return AccessTypePrivate
	default:
		return AccessTypePrivate
	}
}
