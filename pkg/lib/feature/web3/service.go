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
	Web3Config  *config.Web3Config
}

func (s *Service) GetWeb3Info(addresses []string) (*model.UserWeb3Info, error) {
	if s.Web3Config == nil || s.Web3Config.NFT == nil {
		return nil, fmt.Errorf("NFTConfig not defined")
	}
	nftCollections := s.Web3Config.NFT.Collections
	contractIDs := make([]web3.ContractID, 0, len(nftCollections))
	for _, collection := range nftCollections {
		contractID, err := web3.ParseContractID(collection)
		if err != nil {
			return nil, err
		}
		contractIDs = append(contractIDs, *contractID)
	}

	ownerships := make([]model.NFTOwnership, 0)
	for _, address := range addresses {
		nfts, err := s.GetNFTsByAddress(contractIDs, address)
		if err != nil {
			return nil, err
		}

		ownerships = append(ownerships, (nfts.Items)...)

	}

	web3Info := &model.UserWeb3Info{
		Accounts: ownerships,
	}

	return web3Info, nil
}

func (s *Service) GetNFTsByAddress(contracts []web3.ContractID, ownerAddresses string) (*model.GetUserNFTsResponse, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = path.Join("nfts", ownerAddresses)

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
