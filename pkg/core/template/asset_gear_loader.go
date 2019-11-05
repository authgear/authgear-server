package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

	return DownloadStringFromAssuminglyTrustedURL(signedURL)
}
