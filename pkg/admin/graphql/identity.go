package graphql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeIdentity = "Identity"

var nodeIdentity = entity(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeIdentity,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id": entityIDField(typeIdentity, func(obj interface{}) (string, error) {
				ref := obj.(interface{ ToRef() *identity.Ref }).ToRef()
				return encodeIdentityID(ref), nil
			}),
			"createdAt": entityCreatedAtField(loadIdentity),
			"updatedAt": entityUpdatedAtField(loadIdentity),
			"type": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ref := p.Source.(interface{ ToRef() *identity.Ref }).ToRef()
					return string(ref.Type), nil
				},
			},
			"claims": &graphql.Field{
				Type: graphql.NewNonNull(graphqlutil.JSONObject),
				Args: map[string]*graphql.ArgumentConfig{
					"names": {Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					names, hasNames := p.Args["names"].([]interface{})
					info := loadIdentity(p.Context, p.Source)
					claims := info.Map(func(value interface{}) (interface{}, error) {
						claims := value.(*identity.Info).Claims
						if hasNames {
							filteredClaims := make(map[string]interface{})
							for _, name := range names {
								name := name.(string)
								if value, ok := claims[name]; ok {
									filteredClaims[name] = value
								}
							}
							claims = filteredClaims
						}
						return claims, nil
					})
					return claims.Value, nil
				},
			},
		},
	}),
	&identity.Info{},
	func(ctx *Context, id string) (interface{}, error) {
		ref, err := decodeIdentityID(id)
		if err != nil {
			return nil, err
		}

		return ctx.Identities.Get(ref).Value, nil
	},
)

var connIdentity = connection(nodeIdentity)

func encodeIdentityID(ref *identity.Ref) string {
	return strings.Join([]string{
		string(ref.Type),
		ref.ID,
	}, "|")
}

func decodeIdentityID(id string) (*identity.Ref, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return nil, errors.New("invalid ID")
	}
	return &identity.Ref{
		Type: authn.IdentityType(parts[0]),
		Meta: model.Meta{ID: parts[1]},
	}, nil
}

func loadIdentity(ctx context.Context, obj interface{}) *graphqlutil.Lazy {
	switch obj := obj.(type) {
	case *identity.Info:
		return graphqlutil.NewLazyValue(obj)
	case *identity.Ref:
		return GQLContext(ctx).Identities.Get(obj)
	default:
		panic(fmt.Sprintf("graphql: unknown identity type: %T", obj))
	}
}
