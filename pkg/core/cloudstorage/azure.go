package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type AzureStorage struct {
	StorageAccount string
	Container      string
	// AccessKey is encoded in standard BASE64.
	AccessKey string
}

var _ Storage = &AzureStorage{}

func NewAzureStorage(storageAccount string, accessKey string, container string) *AzureStorage {
	return &AzureStorage{
		StorageAccount: storageAccount,
		Container:      container,
		AccessKey:      accessKey,
	}
}

const (
	AzureHeaderAccess = "x-ms-meta-access"
)

var AzureProprietaryToStandardMap = map[string]string{
	// The names have hyphens removed because they must be valid C# identifiers.
	"x-ms-meta-accesscontrolalloworigin":      "access-control-allow-origin",
	"x-ms-meta-accesscontrolexposeheaders":    "access-control-expose-headers",
	"x-ms-meta-accesscontrolmaxage":           "access-control-max-age",
	"x-ms-meta-accesscontrolallowcredentials": "access-control-allow-credentials",
	"x-ms-meta-accesscontrolallowmethods":     "access-control-allow-methods",
	"x-ms-meta-accesscontrolallowheaders":     "access-control-allow-headers",
}

var AzureStandardToProprietaryMap = map[string]string{
	"content-disposition": "x-ms-blob-content-disposition",
	// The names have hyphens removed because they must be valid C# identifiers.
	"access-control-allow-origin":      "x-ms-meta-accesscontrolalloworigin",
	"access-control-expose-headers":    "x-ms-meta-accesscontrolexposeheaders",
	"access-control-max-age":           "x-ms-meta-accesscontrolmaxage",
	"access-control-allow-credentials": "x-ms-meta-accesscontrolallowcredentials",
	"access-control-allow-methods":     "x-ms-meta-accesscontrolallowmethods",
	"access-control-allow-headers":     "x-ms-meta-accesscontrolallowheaders",
}

func (s *AzureStorage) PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error) {
	now := time.Now().UTC()
	u, err := s.SignedURL(name, now, azblob.BlobSASPermissions{
		Create: true,
		Write:  true,
	})
	if err != nil {
		return nil, err
	}

	header = s.StandardToProprietary(header)
	header.Set("x-ms-blob-type", "BlockBlob")
	header.Set(AzureHeaderAccess, string(accessType))

	req := http.Request{
		Method: "PUT",
		Header: header,
		URL:    u,
	}

	return &req, nil
}

func (s *AzureStorage) PresignGetObject(name string) (*url.URL, error) {
	now := time.Now().UTC()
	return s.SignedURL(name, now, azblob.BlobSASPermissions{
		Read: true,
	})
}

func (s *AzureStorage) StandardToProprietary(header http.Header) http.Header {
	return RewriteHeaderName(header, AzureStandardToProprietaryMap)
}

func (s *AzureStorage) ProprietaryToStandard(header http.Header) http.Header {
	return RewriteHeaderName(header, AzureProprietaryToStandardMap)
}

func (s *AzureStorage) AccessType(header http.Header) AccessType {
	a := header.Get(AzureHeaderAccess)
	switch a {
	case string(AccessTypePublic):
		return AccessTypePublic
	case string(AccessTypePrivate):
		return AccessTypePrivate
	default:
		return AccessTypePrivate
	}
}

func (s *AzureStorage) SignedURL(name string, now time.Time, perm azblob.BlobSASPermissions) (*url.URL, error) {
	sigValues := azblob.BlobSASSignatureValues{
		Version:       "2018-11-09",
		Protocol:      azblob.SASProtocolHTTPS,
		StartTime:     now,
		ExpiryTime:    now.Add(1 * time.Hour),
		Permissions:   perm.String(),
		ContainerName: s.Container,
		BlobName:      name,
	}

	cred, err := azblob.NewSharedKeyCredential(s.StorageAccount, s.AccessKey)
	if err != nil {
		return nil, err
	}

	q, err := sigValues.NewSASQueryParameters(cred)
	if err != nil {
		return nil, err
	}

	parts := azblob.BlobURLParts{
		Scheme:        "https",
		Host:          fmt.Sprintf("%s.blob.core.windows.net", s.StorageAccount),
		ContainerName: s.Container,
		BlobName:      name,
		SAS:           q,
	}
	u := parts.URL()

	return &u, nil
}
