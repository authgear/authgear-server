package model

import (
	"math/big"
	"strconv"
	"time"

	web3util "github.com/authgear/authgear-server/pkg/util/web3"
)

type AccountIdentifier struct {
	Address string `json:"address"`
}

type NetworkIdentifier struct {
	Blockchain string `json:"blockchain"`
	Network    string `json:"network"`
}

type Contract struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type TransactionIdentifier struct {
	Hash string `json:"hash"`
}

type BlockIdentifier struct {
	Index     big.Int   `json:"index"`
	Timestamp time.Time `json:"timestamp"`
}

type Token struct {
	TokenID               big.Int               `json:"token_id"`
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
	BlockIdentifier       BlockIdentifier       `json:"block_identifier"`
}

type NFT struct {
	Contract Contract `json:"contract"`
	Balance  int      `json:"balance"`
	Tokens   []Token  `json:"tokens"`
}

type NFTOwnership struct {
	AccountIdentifier AccountIdentifier `json:"account_identifier"`
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	NFTs              []NFT             `json:"nfts"`
}

func (s *NFTOwnership) EndUserAccountID() string {
	if s.NetworkIdentifier.Blockchain == "ethereum" {
		chainID, err := strconv.ParseInt(s.NetworkIdentifier.Network, 10, 0)
		if err != nil {
			return ""
		}
		eip681 := web3util.EIP681{
			ChainID: int(chainID),
			Address: s.AccountIdentifier.Address,
		}
		return eip681.URL().String()
	}

	return ""
}

type NFTCollection struct {
	ID              string    `json:"id"`
	Blockchain      string    `json:"blockchain"`
	Network         string    `json:"network"`
	Name            string    `json:"name"`
	BlockHeight     int       `json:"block_height"`
	ContractAddress string    `json:"contract_address"`
	TotalSupply     int       `json:"total_supply"`
	TokenType       string    `json:"type"`
	CreatedAt       time.Time `json:"created_at"`
}

type ContractMetadataMetadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	TotalSupply string `json:"total_supply"`
	TokenType   string `json:"token_type"`
}

type WatchCollectionRequest struct {
	ContractID string `json:"contract_id"`
	Name       string `json:"name,omitempty"`
}

type GetCollectionsResult struct {
	Items []NFTCollection `json:"items"`
}

type ContractMetadata struct {
	Address  string                   `json:"address"`
	Metadata ContractMetadataMetadata `json:"contract_metadata"`
}

type UserWeb3Info struct {
	Accounts []NFTOwnership `json:"accounts"`
}

type EthereumNetwork string

const (
	EthereumNetworkEthereumMainnet EthereumNetwork = "1"
	EthereumNetworkEthereumGoerli  EthereumNetwork = "5"
	EthereumNetworkPolygonMainnet  EthereumNetwork = "137"
	EthereumNetworkPolygonMumbai   EthereumNetwork = "80001"
)

func ParseEthereumNetwork(s string) (EthereumNetwork, bool) {
	switch s {
	case "1":
		return EthereumNetworkEthereumMainnet, true
	case "5":
		return EthereumNetworkEthereumGoerli, true
	case "137":
		return EthereumNetworkPolygonMainnet, true
	case "80001":
		return EthereumNetworkPolygonMumbai, true
	default:
		return "", false
	}
}
