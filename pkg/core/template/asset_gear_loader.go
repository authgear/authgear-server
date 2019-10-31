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

func (l *AssetGearLoader) Load(u *url.URL) (templateContent string, err error) {
	// The url is asset-gear:///assetname
	path := u.Path
	assetName := strings.TrimPrefix(path, "/")
	reqBody := map[string]interface{}{
		"assets": []interface{}{
			map[string]interface{}{
				"asset_name": assetName,
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

	respBody := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = errors.HandledWithMessage(err, "unexpected response body")
		return
	}

	signedURL, ok := respBody["result"].(map[string]interface{})["assets"].([]interface{})[0].(map[string]interface{})["url"].(string)
	if !ok {
		err = errors.New("failed to get signed template URL")
		return
	}

	return DownloadStringFromAssuminglyTrustedURL(signedURL)
}
