package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AzureStorage struct {
	ServiceURL     string
	StorageAccount string
	Container      string
	AccessKey      string
	Clock          clock.Clock
}

var _ Storage = &AzureStorage{}

func NewAzureStorage(serviceURL string, storageAccount string, accessKey string, container string, c clock.Clock) *AzureStorage {
	return &AzureStorage{
		ServiceURL:     serviceURL,
		StorageAccount: storageAccount,
		Container:      container,
		AccessKey:      accessKey,
		Clock:          c,
	}
}

func (s *AzureStorage) getServiceURL() (*url.URL, error) {
	serviceURL := s.ServiceURL
	if serviceURL == "" {
		serviceURL = fmt.Sprintf("https://%s.blob.core.windows.net", s.StorageAccount)
	}

	u, err := url.Parse(serviceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Azure blob service URL: %w", err)
	}

	return u, nil
}

func (s *AzureStorage) PresignPutObject(name string, header http.Header) (*http.Request, error) {
	now := s.Clock.NowUTC()
	u, err := s.SignedURL(name, now, PresignPutExpires, azblob.BlobSASPermissions{
		Create: true,
		Write:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to presign put request: %w", err)
	}

	header = header.Clone()
	header.Set("x-ms-blob-type", "BlockBlob")

	req := http.Request{
		Method: "PUT",
		Header: header,
		URL:    u,
	}

	return &req, nil
}

func (s *AzureStorage) PresignGetObject(name string) (*url.URL, error) {
	now := s.Clock.NowUTC()
	return s.SignedURL(name, now, PresignGetExpires, azblob.BlobSASPermissions{
		Read: true,
	})
}

func (s *AzureStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.PresignGetObject(name)
}

func (s *AzureStorage) SignedURL(name string, now time.Time, duration time.Duration, perm azblob.BlobSASPermissions) (*url.URL, error) {
	sigValues := azblob.BlobSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,
		StartTime:     now,
		ExpiryTime:    now.Add(duration),
		Permissions:   perm.String(),
		ContainerName: s.Container,
		BlobName:      name,
	}

	cred, err := azblob.NewSharedKeyCredential(s.StorageAccount, s.AccessKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure credentials: %w", err)
	}

	q, err := sigValues.NewSASQueryParameters(cred)
	if err != nil {
		return nil, fmt.Errorf("failed to create SAS query parameters: %w", err)
	}

	serviceURL, err := s.getServiceURL()
	if err != nil {
		return nil, err
	}

	parts := azblob.BlobURLParts{
		Scheme:        serviceURL.Scheme,
		Host:          serviceURL.Host,
		ContainerName: s.Container,
		BlobName:      name,
		SAS:           q,
	}
	u := parts.URL()

	return &u, nil
}

func (s *AzureStorage) MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)
		u, err := s.PresignGetObject(key)
		if err != nil {
			panic(err)
		}
		r.Host = ""
		r.URL = u
	}
}
