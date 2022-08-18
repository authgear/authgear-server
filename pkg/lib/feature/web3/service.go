package web3

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type Service struct {
	APIEndpoint config.NFTIndexerAPIEndpoint
}

func (s *Service) GetWeb3Info(addresses []string) (map[string]interface{}, error) {
	nfts, err := s.GetNFTsByAddresses(addresses)
	if err != nil {
		return nil, err
	}

	web3Info := map[string]interface{}{
		"nfts": nfts.Items,
	}

	return web3Info, nil
}

func (s *Service) WatchNFTCollection(contractID web3.ContractID) (*model.WatchColletionResponse, error) {

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

func (s *Service) GetNFTsByAddresses(addresses []string) (*model.GetUserNFTsResponse, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("nfts")
	endpoint.Query().Set("owner_addresses", strings.Join(addresses, ","))

	res, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}

	var response model.GetUserNFTsResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *Service) GetNFTCollections() (*model.GetCollectionsResponse, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("collections")

	res, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}

	var response model.GetCollectionsResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
