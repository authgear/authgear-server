package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type NFTService struct {
	APIEndpoint config.NFTIndexerAPIEndpoint
}

func (s *NFTService) WatchNFTCollection(contractID web3.ContractID) (*model.WatchColletionResponse, error) {
	if s.APIEndpoint == "" {
		return nil, nil
	}

	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("watch")

	contractURL, err := contractID.URL()
	if err != nil {
		return nil, err
	}

	request := model.WatchCollectionRequest{
		ContractID: contractURL.String(),
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(endpoint.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var response model.WatchColletionResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
