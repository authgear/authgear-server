package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

var probeNFTCollectionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ProbeNFTCollectionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"contractID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var probeNFTCollectionsPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ProbeNFTCollectionsPayload",
	Fields: graphql.Fields{
		"isLargeCollection": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Boolean),
		},
	},
})

var _ = registerMutationField(
	"probeNFTCollection",
	&graphql.Field{
		Description: "Probes a NFT Collection to see whether it is a large collection",
		Type:        graphql.NewNonNull(probeNFTCollectionsPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(probeNFTCollectionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			rawContractURL := input["contractID"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			contractID, err := web3.ParseContractID(rawContractURL)
			if err != nil {
				return nil, err
			}

			res, err := gqlCtx.NFTService.ProbeNFTCollection(*contractID)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"isLargeCollection": res.IsLargeCollection,
			}).Value, nil
		},
	},
)
