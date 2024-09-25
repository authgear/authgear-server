package cloudstorage

import (
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

// PresignPutExpires is how long the presign PUT request remains valid.
const PresignPutExpires time.Duration = 15 * duration.PerMinute

type storage interface {
	// PresignPutObject returns an HTTP request that is ready for use.
	PresignPutObject(name string, header http.Header) (*http.Request, error)
	// PresignHeadObject returns an URL that is ready for use.
	PresignHeadObject(name string, expire time.Duration) (*url.URL, error)
	// PresignGetObject returns an URL that is ready for use.
	PresignGetObject(name string, expire time.Duration) (*url.URL, error)
	// MakeDirector takes extractKey and returns a Director of httputil.ReverseProxy.
	MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request)
}
