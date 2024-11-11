package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type ProbeColletionResponse struct {
	Result apimodel.ProbeCollectionResult `json:"result"`
	Error  *apierrors.APIError            `json:"error"`
}

type GetContractMetadataResponse struct {
	Result apimodel.GetContractMetadataResult `json:"result"`
	Error  *apierrors.APIError                `json:"error"`
}

type NFTService struct {
	HTTPClient  HTTPClient
	APIEndpoint config.NFTIndexerAPIEndpoint
}

func (s *NFTService) ProbeNFTCollection(ctx context.Context, contractID web3.ContractID) (*apimodel.ProbeCollectionResult, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = "probe"

	contractURL, err := contractID.URL()
	if err != nil {
		return nil, err
	}

	request := model.ProbeCollectionRequest{
		ContractID: contractURL.String(),
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	res, err := httputil.PostWithContext(ctx, s.HTTPClient.Client, endpoint.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response ProbeColletionResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return &response.Result, nil
}

func (s *NFTService) GetContractMetadata(ctx context.Context, contracts []web3.ContractID) ([]apimodel.NFTCollection, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = "metadata"

	contractURLs := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		contractURL, err := contract.URL()
		if err != nil {
			return nil, err
		}
		contractURLs = append(contractURLs, contractURL.String())
	}

	request := model.GetContractMetadataRequest{
		ContractIDs: contractURLs,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	res, err := httputil.PostWithContext(ctx, s.HTTPClient.Client, endpoint.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response GetContractMetadataResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result.Collections, nil
}
