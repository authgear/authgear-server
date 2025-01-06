package httputil

// We used to use https://github.com/nytimes/gziphandler to as a gzip middleware.
// But go 1.23 explicitly states that ServeContent and friends does not support Content-Encoding: gzip.
// So we removed this gzip middleware until we find an alternative that uses Transfer-Encoding instead.
// As of 2025-01-06, no such alternative exists.
