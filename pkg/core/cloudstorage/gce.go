package cloudstorage

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type GCEStorage struct {
	ServiceAccount string
	Bucket         string
	// PrivateKey is PEM.
	PrivateKey []byte
}

var _ Storage = &GCEStorage{}

func NewGCEStorage(serviceAccount string, privateKey []byte, bucket string) *GCEStorage {
	return &GCEStorage{
		ServiceAccount: serviceAccount,
		Bucket:         bucket,
		PrivateKey:     privateKey,
	}
}

const (
	GCEHeaderAccess = "x-goog-meta-access"
)

var GCEProprietaryToStandardMap = map[string]string{
	"x-goog-meta-accesscontrolalloworigin":      "access-control-allow-origin",
	"x-goog-meta-accesscontrolexposeheaders":    "access-control-expose-headers",
	"x-goog-meta-accesscontrolmaxage":           "access-control-max-age",
	"x-goog-meta-accesscontrolallowcredentials": "access-control-allow-credentials",
	"x-goog-meta-accesscontrolallowmethods":     "access-control-allow-methods",
	"x-goog-meta-accesscontrolallowheaders":     "access-control-allow-headers",
}

var GCEStandardToProprietaryMap = map[string]string{
	"access-control-allow-origin":      "x-goog-meta-accesscontrolalloworigin",
	"access-control-expose-headers":    "x-goog-meta-accesscontrolexposeheaders",
	"access-control-max-age":           "x-goog-meta-accesscontrolmaxage",
	"access-control-allow-credentials": "x-goog-meta-accesscontrolallowcredentials",
	"access-control-allow-methods":     "x-goog-meta-accesscontrolallowmethods",
	"access-control-allow-headers":     "x-goog-meta-accesscontrolallowheaders",
}

func (s *GCEStorage) PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error) {
	now := time.Now().UTC()

	header = s.StandardToProprietary(header)
	header.Set(GCEHeaderAccess, string(accessType))

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

func (s *GCEStorage) PresignGetOrHeadObject(name string, method string) (*url.URL, error) {
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

func (s *GCEStorage) PresignGetObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "GET")
}

func (s *GCEStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "HEAD")
}

func (s *GCEStorage) StandardToProprietary(header http.Header) http.Header {
	return RewriteHeaderName(header, GCEStandardToProprietaryMap)
}

func (s *GCEStorage) ProprietaryToStandard(header http.Header) http.Header {
	return RewriteHeaderName(header, GCEProprietaryToStandardMap)
}

func (s *GCEStorage) AccessType(header http.Header) AccessType {
	a := header.Get(GCEHeaderAccess)
	switch a {
	case string(AccessTypePublic):
		return AccessTypePublic
	case string(AccessTypePrivate):
		return AccessTypePrivate
	default:
		return AccessTypePrivate
	}
}
