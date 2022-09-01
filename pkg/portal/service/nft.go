package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type WatchColletionResponse struct {
	Result apimodel.NFTCollection `json:"result"`
	Error  *apierrors.APIError    `json:"error"`
}

type GetContractMetadataResponse struct {
	Result apimodel.ContractMetadata `json:"result"`
	Error  *apierrors.APIError       `json:"error"`
}

type GetCollectionsResponse struct {
	Result apimodel.GetCollectionsResult `json:"result"`
	Error  *apierrors.APIError           `json:"error"`
}

type NFTService struct {
	APIEndpoint config.NFTIndexerAPIEndpoint
}

func (s *NFTService) WatchNFTCollection(contractID web3.ContractID) (*apimodel.NFTCollection, error) {
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

	var response WatchColletionResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return &response.Result, nil
}

func (s *NFTService) GetNFTCollections(contracts []web3.ContractID) (*apimodel.GetCollectionsResult, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("collections")

	query := endpoint.Query()
	if len(contracts) > 0 {
		urls := make([]string, 0, len(contracts))
		for _, contract := range contracts {
			url, err := contract.URL()
			if err != nil {
				return nil, err
			}
			urls = append(urls, url.String())
		}
		query["contract_id"] = urls
	}

	endpoint.RawQuery = query.Encode()

	res, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}

	var response GetCollectionsResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return &response.Result, nil
}

func (s *NFTService) GetContractMetadata(appID string, contract web3.ContractID) (*apimodel.ContractMetadata, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	contractURL, err := contract.URL()
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("metadata", contractURL.String())

	query := endpoint.Query()
	query.Set("app_id", appID)

	endpoint.RawQuery = query.Encode()

	res, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}

	var response GetContractMetadataResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return &response.Result, nil
}
