package graphql

import (
	"context"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeAuthenticator = "Authenticator"

var authenticatorType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AuthenticatorType",
	Values: graphql.EnumValueConfigMap{
		"PASSWORD": &graphql.EnumValueConfig{
			Value: "password",
		},
		"TOTP": &graphql.EnumValueConfig{
			Value: "totp",
		},
		"OOB_OTP": &graphql.EnumValueConfig{
			Value: "oob_otp",
		},
	},
})

var authenticatorKind = graphql.NewEnum(graphql.EnumConfig{
	Name: "AuthenticatorKind",
	Values: graphql.EnumValueConfigMap{
		"PRIMARY": &graphql.EnumValueConfig{
			Value: "primary",
		},
		"SECONDARY": &graphql.EnumValueConfig{
			Value: "secondary",
		},
	},
})

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
				Type: graphql.NewNonNull(authenticatorType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ref := p.Source.(interface{ ToRef() *authenticator.Ref }).ToRef()
					return string(ref.Type), nil
				},
			},
			"kind": &graphql.Field{
				Type: graphql.NewNonNull(authenticatorKind),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					info := loadAuthenticator(p.Context, p.Source)
					return info.Map(func(value interface{}) (interface{}, error) {
						a := value.(*authenticator.Info)
						return string(a.Kind), nil
					}).Value, nil
				},
			},
			"isDefault": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					info := loadAuthenticator(p.Context, p.Source)
					return info.Map(func(value interface{}) (interface{}, error) {
						a := value.(*authenticator.Info)
						return a.IsDefault, nil
					}).Value, nil
				},
			},
			"claims": &graphql.Field{
				Type: graphql.NewNonNull(AuthenticatorClaims),
				Args: map[string]*graphql.ArgumentConfig{
					"names": {Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					names, hasNames := p.Args["names"].([]interface{})
					info := loadAuthenticator(p.Context, p.Source)
					claims := info.Map(func(value interface{}) (interface{}, error) {
						claims := value.(*authenticator.Info).Claims
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
	&authenticator.Info{},
	func(ctx *Context, id string) (interface{}, error) {
		ref, err := decodeAuthenticatorID(id)
		if err != nil {
			return nil, err
		}

		return ctx.Authenticators.Get(ref).Value, nil
	},
)

var connAuthenticator = graphqlutil.NewConnectionDef(nodeAuthenticator)

func encodeAuthenticatorID(ref *authenticator.Ref) string {
	return strings.Join([]string{
		string(ref.Type),
		ref.ID,
	}, "|")
}

func decodeAuthenticatorID(id string) (*authenticator.Ref, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return nil, apierrors.NewInvalid("invalid ID")
	}
	return &authenticator.Ref{
		Type: authn.AuthenticatorType(parts[0]),
		Meta: model.Meta{ID: parts[1]},
	}, nil
}

func loadAuthenticator(ctx context.Context, obj interface{}) *graphqlutil.Lazy {
	switch obj := obj.(type) {
	case *authenticator.Info:
		return graphqlutil.NewLazyValue(obj)
	case *authenticator.Ref:
		return GQLContext(ctx).Authenticators.Get(obj)
	default:
		panic(fmt.Sprintf("graphql: unknown authenticator type: %T", obj))
	}
}
