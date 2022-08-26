package model

import (
	"math/big"
	"time"
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

type NFTCollection struct {
	ID              string `json:"id"`
	Blockchain      string `json:"blockchain"`
	Network         string `json:"network"`
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
}

type GetUserNFTsResponse struct {
	Items []NFTOwnership `json:"items"`
}

type WatchCollectionRequest struct {
	ContractID string `json:"contract_id"`
	Name       string `json:"name,omitempty"`
}

type WatchColletionResponse struct {
	ID              string `json:"id"`
	Blockchain      string `json:"blockchain"`
	Network         string `json:"network"`
	ContractAddress string `json:"contract_address"`
	Name            string `json:"name,omitempty"`
}

type GetCollectionsResponse struct {
	Items []NFTCollection `json:"items"`
}

type UserWeb3Info struct {
	Accounts []NFTOwnership `json:"accounts"`
}

type EthereumNetwork string

const (
	EthereumNetworkEthereumMainnet EthereumNetwork = "1"
)

func ParseEthereumNetwork(s string) (EthereumNetwork, bool) {
	switch s {
	case "1":
		return EthereumNetworkEthereumMainnet, true
	default:
		return "", false
	}
}
