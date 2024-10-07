package e2eclient

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SAMLClient struct {
	Context    context.Context
	HTTPClient *http.Client
	HTTPHost   httputil.HTTPHost
}

func (c *SAMLClient) SendSAMLRequestWithHTTPRedirect(
	samlRequestXML string,
	destination *url.URL,
	onResponse func(r *http.Response) error) error {
	compressedRequestBuffer := &bytes.Buffer{}
	writer, err := flate.NewWriter(compressedRequestBuffer, 9)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(samlRequestXML))
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	base64EncodedRequest := base64.StdEncoding.EncodeToString(compressedRequestBuffer.Bytes())
	q := &url.Values{
		"SAMLRequest": []string{base64EncodedRequest},
	}
	u := destination
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(c.Context, "GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Host = string(c.HTTPHost)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return onResponse(resp)
}

func (c *SAMLClient) SendSAMLRequestWithHTTPPost(
	samlRequestXML string,
	destination *url.URL,
	onResponse func(r *http.Response) error) error {
	base64EncodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequestXML))
	body := &url.Values{
		"SAMLRequest": []string{base64EncodedRequest},
	}
	bodyBuffer := bytes.NewBuffer([]byte(body.Encode()))
	u := destination
	req, err := http.NewRequestWithContext(c.Context, "POST", u.String(), bodyBuffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = string(c.HTTPHost)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return onResponse(resp)
}
