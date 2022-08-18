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
	Index big.Int `json:"index"`
}

type NFTOwner struct {
	AccountIdentifier     AccountIdentifier     `json:"account_identifier"`
	NetworkIdentifier     NetworkIdentifier     `json:"network_identifier"`
	Contract              Contract              `json:"contract"`
	TokenID               big.Int               `json:"token_id"`
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
	BlockIdentifier       BlockIdentifier       `json:"block_identifier"`
	Timestamp             time.Time             `json:"timestamp"`
}

type NFTCollection struct {
	ID              string `json:"id"`
	Blockchain      string `json:"blockchain"`
	Network         string `json:"network"`
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
}

type GetUserNFTsResponse struct {
	Items      []NFTOwner `json:"items"`
	TotalCount int        `json:"total_count"`
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
	NFTs []NFTOwner `json:"nfts"`
}
