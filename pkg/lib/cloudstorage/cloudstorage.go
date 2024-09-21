package cloudstorage

import (
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

// PresignPutExpires is how long the presign PUT request remains valid.
const PresignPutExpires time.Duration = 15 * duration.PerMinute

// PresignGetExpires is how long the presign GET request remains valid.
const PresignGetExpires time.Duration = 1 * duration.PerHour

// PresignGetExpiresForUserExport is how long the presign GET request remains valid for user export.
const PresignGetExpiresForUserExport time.Duration = 1 * duration.PerMinute

type Storage interface {
	// PresignPutObject returns an HTTP request that is ready for use.
	PresignPutObject(name string, header http.Header) (*http.Request, error)
	// PresignHeadObject returns an URL that is ready for use.
	PresignHeadObject(name string, expire time.Duration) (*url.URL, error)
	// MakeDirector takes extractKey and returns a Director of httputil.ReverseProxy.
	MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request)
}
