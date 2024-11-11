package web3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

type Service struct {
	APIEndpoint config.NFTIndexerAPIEndpoint
	Web3Config  *config.Web3Config
	HTTPClient  HTTPClient
}

type ListOwnerNFTsResponse struct {
	Result model.NFTOwnership  `json:"result"`
	Error  *apierrors.APIError `json:"error"`
}

func (s *Service) GetWeb3Info(ctx context.Context, identities []*identity.Info) (*model.UserWeb3Info, error) {
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
	for _, identity := range identities {
		if identity == nil {
			continue
		}
		var ownerID *web3.ContractID
		switch identity.Type {
		case model.IdentityTypeSIWE:
			id, err := identity.SIWE.ToContractID()
			if err != nil {
				return nil, err
			}

			ownerID = id
			break

		default:
			// No supported identities
			break
		}

		if ownerID == nil {
			continue
		}

		nft, err := s.ListOwnerNFTs(ctx, *ownerID, contractIDs)
		if err != nil {
			return nil, err
		}

		if nft == nil {
			return nil, fmt.Errorf("Failed to fetch nfts for user")
		}

		ownerships = append(ownerships, *nft)

	}

	web3Info := &model.UserWeb3Info{
		Accounts: ownerships,
	}

	return web3Info, nil
}

func (s *Service) ListOwnerNFTs(ctx context.Context, ownerID web3.ContractID, contractIDs []web3.ContractID) (*model.NFTOwnership, error) {
	endpoint, err := url.Parse(string(s.APIEndpoint))
	if err != nil {
		return nil, err
	}

	endpoint.Path = "nfts"

	ownerURL, err := ownerID.URL()
	if err != nil {
		return nil, err
	}

	contractUrls := make([]string, 0, len(contractIDs))
	for _, contract := range contractIDs {
		url, err := contract.URL()
		if err != nil {
			return nil, err
		}
		contractUrls = append(contractUrls, url.String())
	}

	request := model.ListOwnerNFTsRequest{
		OwnerAddress: ownerURL.String(),
		ContractIDs:  contractUrls,
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

	var response ListOwnerNFTsResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return &response.Result, nil
}
