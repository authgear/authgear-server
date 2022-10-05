package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
)

const typeNFTCollection = "NFTCollection"
const typeNFTContractMetadata = "NFTContractMetadata"

var nftCollection = graphql.NewObject(graphql.ObjectConfig{
	Name:        typeNFTCollection,
	Description: "Web3 NFT Collection",
	Fields: graphql.Fields{
		"blockchain": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"network": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"name": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"contractAddress": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"blockHeight": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(model.NFTCollection)

				return source.BlockHeight.String(), nil
			},
		},
		"totalSupply": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(model.NFTCollection)

				if source.TotalSupply == nil {
					return nil, nil
				}

				return source.TotalSupply.String(), nil
			},
		},
		"tokenType": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"createdAt": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
		},
	},
})

var nftContractMetadata = graphql.NewObject(graphql.ObjectConfig{
	Name:        typeNFTContractMetadata,
	Description: "Web3 NFT ContractMetadata",
	Fields: graphql.Fields{
		"address": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"name": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*model.ContractMetadata)

				return source.Metadata.Name, nil
			},
		},
		"symbol": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*model.ContractMetadata)

				return source.Metadata.Symbol, nil
			},
		},
		"totalSupply": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*model.ContractMetadata)

				return source.Metadata.TotalSupply, nil
			},
		},
		"tokenType": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*model.ContractMetadata)

				return source.Metadata.TokenType, nil
			},
		},
	},
})
