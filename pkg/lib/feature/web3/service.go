package web3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type Service struct {
	APIEndpoint config.NFTIndexerAPIEndpoint
	NFTConfig   *config.NFTConfig
}

func (s *Service) GetWeb3Info(addresses []string) (*model.UserWeb3Info, error) {
	if s.NFTConfig == nil {
		return nil, fmt.Errorf("NFTConfig not defined")
	}
	nftCollections := s.NFTConfig.Collections
	contractIDs := make([]web3.ContractID, 0, len(nftCollections))
	for _, collection := range nftCollections {
		contractID, err := web3.ParseContractID(collection)
		if err != nil {
			return nil, err
		}
		contractIDs = append(contractIDs, *contractID)
	}

	nfts, err := s.GetNFTsByAddresses(contractIDs, addresses)
	if err != nil {
		return nil, err
	}

	web3Info := &model.UserWeb3Info{
		NFTs: nfts.Items,
	}

	return web3Info, nil
}

func (s *Service) GetNFTsByAddresses(contracts []web3.ContractID, ownerAddresses []string) (*model.GetUserNFTsResponse, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("nfts")

	query := endpoint.Query()
	if len(ownerAddresses) > 0 {
		query["owner_address"] = ownerAddresses
	}

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
