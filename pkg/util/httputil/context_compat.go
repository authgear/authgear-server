package httputil

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GetWithContext is a compat method for http.Client.Get
func GetWithContext(ctx context.Context, c *http.Client, url string) (resp *http.Response, err error) {
	// This implementation is copied from the stdlib, except that http.NewRequest is replaced with http.NewRequestWithContext.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	return c.Do(req)
}

// HeadWithContext is a compat method for http.Client.Head
func HeadWithContext(ctx context.Context, c *http.Client, url string) (resp *http.Response, err error) {
	// This implementation is copied from the stdlib, except that http.NewRequest is replaced with http.NewRequestWithContext.
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return
	}
	return c.Do(req)
}

// PostWithContext is a compat method for http.Client.Post
func PostWithContext(ctx context.Context, c *http.Client, url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	// This implementation is copied from the stdlib, except that http.NewRequest is replaced with http.NewRequestWithContext.
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// PostFormWithContext is a compat method for http.Client.PostForm
func PostFormWithContext(ctx context.Context, c *http.Client, url string, data url.Values) (resp *http.Response, err error) {
	// This implementation is copied from the stdlib, except that we call PostWithContext instead.
	return PostWithContext(ctx, c, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}
