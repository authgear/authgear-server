package graphql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/utils"
	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

const typeAuthenticator = "Authenticator"

var nodeAuthenticator = entity(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeAuthenticator,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id": entityIDField(typeAuthenticator, func(obj interface{}) (string, error) {
				ref := obj.(interface{ ToRef() *authenticator.Ref }).ToRef()
				return encodeAuthenticatorID(ref), nil
			}),
			"createdAt": entityCreatedAtField(loadAuthenticator),
			"updatedAt": entityUpdatedAtField(loadAuthenticator),
			"type": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ref := p.Source.(interface{ ToRef() *authenticator.Ref }).ToRef()
					return string(ref.Type), nil
				},
			},
			"props": &graphql.Field{
				Type: graphql.NewNonNull(JSONObject),
				Args: map[string]*graphql.ArgumentConfig{
					"names": {Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					names, hasNames := p.Args["names"].([]interface{})
					info := loadAuthenticator(p.Context, p.Source)
					claims := info.Map(func(value interface{}) (interface{}, error) {
						props := value.(*authenticator.Info).Props
						if hasNames {
							filteredProps := make(map[string]interface{})
							for _, name := range names {
								name := name.(string)
								if value, ok := props[name]; ok {
									filteredProps[name] = value
								}
							}
							props = filteredProps
						}
						return props, nil
					})
					return claims.Value, nil
				},
			},
		},
	}),
	&authenticator.Info{},
	func(ctx *Context, id string) (interface{}, error) {
		ref, err := decodeAuthenticatorID(id)
		if err != nil {
			return nil, err
		}

		return ctx.Authenticators.Get(ref).Value, nil
	},
)

var connAuthenticator = connection(nodeAuthenticator)

func encodeAuthenticatorID(ref *authenticator.Ref) string {
	return strings.Join([]string{
		string(ref.Type),
		ref.ID,
	}, "|")
}

func decodeAuthenticatorID(id string) (*authenticator.Ref, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return nil, errors.New("invalid ID")
	}
	return &authenticator.Ref{
		Type: authn.AuthenticatorType(parts[0]),
		Meta: model.Meta{ID: parts[1]},
	}, nil
}

func loadAuthenticator(ctx context.Context, obj interface{}) *utils.Lazy {
	switch obj := obj.(type) {
	case *authenticator.Info:
		return utils.NewLazyValue(obj)
	case *authenticator.Ref:
		return GQLContext(ctx).Authenticators.Get(obj)
	default:
		panic(fmt.Sprintf("graphql: unknown authenticator type: %T", obj))
	}
}
