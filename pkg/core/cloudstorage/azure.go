package cloudstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/skygeario/skygear-server/pkg/core/errors"
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
	u, err := s.SignedURL(name, now, PresignPutExpires, azblob.BlobSASPermissions{
		Create: true,
		Write:  true,
	})
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to presign put request")
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
	return s.SignedURL(name, now, PresignGetExpires, azblob.BlobSASPermissions{
		Read: true,
	})
}

func (s *AzureStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.PresignGetObject(name)
}

func (s *AzureStorage) ListObjects(r *ListObjectsRequest) (*ListObjectsResponse, error) {
	ctx := context.Background()

	cred, err := azblob.NewSharedKeyCredential(s.StorageAccount, s.AccessKey)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to create azure credentials")
	}

	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})

	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", s.StorageAccount))
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to parse storage account")
	}

	serviceURL := azblob.NewServiceURL(*u, p)
	containerURL := serviceURL.NewContainerURL(s.Container)

	marker := azblob.Marker{}
	if r.PaginationToken != "" {
		v := r.PaginationToken
		marker.Val = &v
	}

	opts := azblob.ListBlobsSegmentOptions{
		Details: azblob.BlobListingDetails{
			Copy:             false,
			Deleted:          false,
			UncommittedBlobs: false,
			Metadata:         false,
		},
		Prefix:     r.Prefix,
		MaxResults: int32(r.PageSize),
	}

	output, err := containerURL.ListBlobsFlatSegment(ctx, marker, opts)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list objects")
	}

	resp := &ListObjectsResponse{}

	if output.NextMarker.Val != nil && *output.NextMarker.Val != "" {
		resp.PaginationToken = *output.NextMarker.Val
	}

	for _, blob := range output.Segment.BlobItems {
		resp.Assets = append(resp.Assets, AssetItem{
			AssetName: blob.Name,
			Size:      *blob.Properties.ContentLength,
		})
	}
	if resp.Assets == nil {
		resp.Assets = []AssetItem{}
	}

	return resp, nil
}

func (s *AzureStorage) DeleteObject(name string) error {
	ctx := context.Background()

	cred, err := azblob.NewSharedKeyCredential(s.StorageAccount, s.AccessKey)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create azure credentials")
	}

	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})

	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", s.StorageAccount))
	if err != nil {
		return errors.HandledWithMessage(err, "failed to parse storage account")
	}

	serviceURL := azblob.NewServiceURL(*u, p)
	containerURL := serviceURL.NewContainerURL(s.Container)
	blobURL := containerURL.NewBlobURL(name)

	_, err = blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})

	if serr, ok := err.(azblob.StorageError); ok && serr.ServiceCode() == azblob.ServiceCodeBlobNotFound {
		return nil
	}

	if err != nil {
		return errors.HandledWithMessage(err, "failed to delete object")
	}

	return nil
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

func (s *AzureStorage) SignedURL(name string, now time.Time, duration time.Duration, perm azblob.BlobSASPermissions) (*url.URL, error) {
	sigValues := azblob.BlobSASSignatureValues{
		Version:       "2018-11-09",
		Protocol:      azblob.SASProtocolHTTPS,
		StartTime:     now,
		ExpiryTime:    now.Add(duration),
		Permissions:   perm.String(),
		ContainerName: s.Container,
		BlobName:      name,
	}

	cred, err := azblob.NewSharedKeyCredential(s.StorageAccount, s.AccessKey)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to create azure credentials")
	}

	q, err := sigValues.NewSASQueryParameters(cred)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to create SAS query parameters")
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
