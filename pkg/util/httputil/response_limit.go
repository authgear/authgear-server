package httputil

import (
	"errors"
	"io"
	"net/http"
)

// MaxResponseBytes caps the response body of every fetch of a URL a project
// supplied. 2 MiB is orders of magnitude above any legitimate response on
// these paths -- a webhook result, an OIDC discovery document, a JWK set, an
// SMS gateway's reply -- and low enough that a hostile destination cannot
// exhaust memory.
//
// Without it the reads on these paths are unbounded: a project admin could
// point a blocking hook at a host that streams indefinitely and take the
// server down, no address rule involved. The cap counts DECOMPRESSED bytes,
// because it is applied above the transport's transparent gzip handling; a
// small compressed body that expands without bound is refused too.
//
// A fetch that needs a tighter bound still enforces its own: the CIMD
// document limit of 5120 bytes is set by the OAuth spec, not by this.
const MaxResponseBytes int64 = 2 * 1024 * 1024

// ErrResponseTooLarge is returned by a read against a response body that
// exceeds MaxResponseBytes. It surfaces wherever the caller reads the body --
// io.ReadAll, a JSON decode, a schema validation -- so a caller that already
// handles a malformed response handles this too.
var ErrResponseTooLarge = errors.New("httputil: response exceeds the maximum size")

// limitedResponseBodyRoundTripper caps every response body it returns.
//
// It sits in the transport rather than at each place that reads a body, for
// the same reason the refusal log does: a new caller of
// NewSSRFSafeExternalClient cannot forget it, because there is nothing to
// remember.
type limitedResponseBodyRoundTripper struct {
	max  int64
	base http.RoundTripper
}

func (rt *limitedResponseBodyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}
	resp.Body = &limitedResponseBody{
		base: resp.Body,
		// One byte more than the limit, so that a body of exactly max bytes
		// is accepted and the first byte beyond it is what trips the error.
		remaining: rt.max + 1,
	}
	return resp, nil
}

type limitedResponseBody struct {
	base      io.ReadCloser
	remaining int64
}

func (b *limitedResponseBody) Read(p []byte) (int, error) {
	if b.remaining <= 0 {
		return 0, ErrResponseTooLarge
	}
	if int64(len(p)) > b.remaining {
		p = p[:b.remaining]
	}
	n, err := b.base.Read(p)
	b.remaining -= int64(n)
	// Reaching zero means the extra byte was delivered, so the body is over
	// the limit whatever the underlying reader reports -- including io.EOF,
	// which would otherwise present an over-limit body as a complete one.
	if b.remaining <= 0 {
		return n, ErrResponseTooLarge
	}
	return n, err
}

func (b *limitedResponseBody) Close() error {
	return b.base.Close()
}
