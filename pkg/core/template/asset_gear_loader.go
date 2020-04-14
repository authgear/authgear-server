package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type AssetGearLoader struct {
	AssetGearEndpoint  string
	AssetGearMasterKey string
}

type signAssetRequest struct {
	Assets []signAssetItem `json:"assets,omitempty"`
}

type signAssetItem struct {
	AssetName string `json:"asset_name,omitempty"`
	URL       string `json:"url,omitempty"`
}

type responseBody struct {
	Result signAssetRequest `json:"result,omitempty"`
}

func (l *AssetGearLoader) Load(u *url.URL) (templateContent string, err error) {
	// The url is asset-gear:///assetname
	path := u.Path
	assetName := strings.TrimPrefix(path, "/")
	reqBody := signAssetRequest{
		Assets: []signAssetItem{
			signAssetItem{
				AssetName: assetName,
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to encode body")
		return
	}

	signReq, err := http.NewRequest("POST", fmt.Sprintf("%s/_asset/get_signed_url", l.AssetGearEndpoint), bytes.NewReader(body))
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to create sign request")
		return
	}
	signReq.Header.Set("x-skygear-api-key", l.AssetGearMasterKey)
	signReq.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(signReq)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get signed template URL")
		return
	}
	defer resp.Body.Close()

	var respBody responseBody
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = errors.HandledWithMessage(err, "unexpected response body")
		return
	}

	if len(respBody.Result.Assets) <= 0 {
		err = errors.New("failed to get signed template URL")
		return
	}
	signedURL := respBody.Result.Assets[0].URL
	if signedURL == "" {
		err = errors.New("failed to get signed template URL")
		return
	}

	return downloadStringFromAssuminglyTrustedURL(signedURL)
}

// downloadStringFromAssuminglyTrustedURL downloads the content of url.
// url is assumed to be trusted.
func downloadStringFromAssuminglyTrustedURL(url string) (content string, err error) {
	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.Newf("unexpected status code: %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, MaxTemplateSize))
	if err != nil {
		return
	}

	if !utf8.Valid(body) {
		err = errors.New("expected content to be UTF-8 encoded")
		return
	}

	content = string(body)
	return
}
