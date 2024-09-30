package e2eclient

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
)

type SAMLClient struct {
	Context    context.Context
	HTTPClient *http.Client
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
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return onResponse(resp)
}
