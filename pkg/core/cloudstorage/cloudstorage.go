package cloudstorage

import (
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

// AccessType is either public or private.
type AccessType string

const (
	// AccessTypePublic is public access.
	AccessTypePublic AccessType = "public"
	// AccessTypePrivate requires signature.
	AccessTypePrivate AccessType = "private"
)

// Storage is abstraction over various cloud storage providers.
type Storage interface {
	// PresignPutObject returns an HTTP request that is ready for use.
	PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error)
	// PresignGetObject returns an URL that is ready for use.
	PresignGetObject(name string) (*url.URL, error)
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
