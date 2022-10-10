package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
)

const typeNFTCollection = "NFTCollection"

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
