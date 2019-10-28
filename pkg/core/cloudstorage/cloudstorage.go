package cloudstorage

import (
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

// AccessType is either public or private.
type AccessType string

const (
	// AccessTypePublic is public access.
	AccessTypePublic AccessType = "public"
	// AccessTypePrivate requires signature.
	AccessTypePrivate AccessType = "private"
)

// AccessTypeDefault is public.
const AccessTypeDefault = AccessTypePublic

// PresignPutExpires is how long the presign PUT request remains valid.
const PresignPutExpires time.Duration = 15 * time.Minute

// PresignGetExpires is how long the presign GET request remains valid.
const PresignGetExpires time.Duration = 1 * time.Hour

// Storage is abstraction over various cloud storage providers.
type Storage interface {
	// PresignPutObject returns an HTTP request that is ready for use.
	PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error)
	// PresignGetObject returns an URL that is ready for use.
	PresignGetObject(name string) (*url.URL, error)
	// PresignHeadObject returns an URL that is ready for use.
	PresignHeadObject(name string) (*url.URL, error)
	// RewriteGetURL rewrite the given URL so that it is ready for use.
	// If the URL is not originally signed, the returned URL is signed.
	// The second return value indicates whether the original URL is signed or not.
	RewriteGetURL(u *url.URL, name string) (*url.URL, bool, error)
	// ListObjects lists objects in a paginated fashion.
	ListObjects(r *ListObjectsRequest) (*ListObjectsResponse, error)
	// DeleteObject deletes the given object.
	// It is not an error if the object does not exist.
	// This is due to the limitation of S3, which treats
	// deleting non-existent object as success.
	DeleteObject(name string) error
	// AccessType returns AccessType stored in the header.
	AccessType(header http.Header) AccessType
	// StandardToProprietary rewrites any standard headers to proprietary ones
	// so that they can be retrieved later.
	StandardToProprietary(header http.Header) http.Header
	// ProprietaryToStandard does the reverse of StandardToProprietary.
	ProprietaryToStandard(header http.Header) http.Header
}

func RewriteHeaderName(header http.Header, mapping map[string]string) http.Header {
	output := http.Header{}
	for oldName, newName := range mapping {
		newName = textproto.CanonicalMIMEHeaderKey(newName)
		for originalName, value := range header {
			lowerName := strings.ToLower(originalName)
			if oldName == lowerName {
				output[newName] = value
			} else {
				output[originalName] = value
			}
		}
	}
	return output
}
