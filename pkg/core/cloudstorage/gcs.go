package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	ServiceAccount string
	Bucket         string
	// PrivateKey is PEM.
	PrivateKey []byte
}

var _ Storage = &GCSStorage{}

func NewGCSStorage(serviceAccount string, privateKey []byte, bucket string) *GCSStorage {
	return &GCSStorage{
		ServiceAccount: serviceAccount,
		Bucket:         bucket,
		PrivateKey:     privateKey,
	}
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
	now := time.Now().UTC()

	header = s.StandardToProprietary(header)
	header.Set(GCSHeaderAccess, string(accessType))

	// We must omit Content-type and Content-MD5 from header because they are special.
	var headerNames []string
	for name := range header {
		lower := strings.ToLower(name)
		if lower == "content-type" || lower == "content-md5" {
			continue
		}
		headerNames = append(headerNames, name)
	}

	expires := now.Add(1 * time.Hour)
	opts := storage.SignedURLOptions{
		GoogleAccessID: s.ServiceAccount,
		PrivateKey:     s.PrivateKey,
		Method:         "PUT",
		Expires:        expires,
		ContentType:    header.Get("Content-Type"),
		Headers:        headerNames,
		MD5:            header.Get("Content-MD5"),
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := storage.SignedURL(s.Bucket, name, &opts)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req := http.Request{
		Method: "PUT",
		Header: header,
		URL:    u,
	}

	return &req, nil
}

func (s *GCSStorage) PresignGetOrHeadObject(name string, method string) (*url.URL, error) {
	now := time.Now().UTC()
	expires := now.Add(1 * time.Hour)

	opts := storage.SignedURLOptions{
		GoogleAccessID: s.ServiceAccount,
		PrivateKey:     s.PrivateKey,
		Method:         method,
		Expires:        expires,
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := storage.SignedURL(s.Bucket, name, &opts)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *GCSStorage) PresignGetObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "GET")
}

func (s *GCSStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "HEAD")
}

func (s *GCSStorage) RewriteGetURL(u *url.URL, name string) (*url.URL, bool, error) {
	q := u.Query()
	_, hasSignature := q["X-Goog-Signature"]

	if hasSignature {
		rewritten := &url.URL{
			Scheme:   "https",
			Host:     "storage.googleapis.com",
			Path:     fmt.Sprintf("/%s/%s", s.Bucket, name),
			RawQuery: u.RawQuery,
		}
		return rewritten, true, nil
	}

	newlySigned, err := s.PresignGetObject(name)
	return newlySigned, false, err
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
