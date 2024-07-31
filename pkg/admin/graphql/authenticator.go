package graphql

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeAuthenticator = "Authenticator"

var authenticatorType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AuthenticatorType",
	Values: graphql.EnumValueConfigMap{
		"PASSWORD": &graphql.EnumValueConfig{
			Value: string(model.AuthenticatorTypePassword),
		},
		"TOTP": &graphql.EnumValueConfig{
			Value: string(model.AuthenticatorTypeTOTP),
		},
		"OOB_OTP_EMAIL": &graphql.EnumValueConfig{
			Value: string(model.AuthenticatorTypeOOBEmail),
		},
		"OOB_OTP_SMS": &graphql.EnumValueConfig{
			Value: string(model.AuthenticatorTypeOOBSMS),
		},
		"PASSKEY": &graphql.EnumValueConfig{
			Value: string(model.AuthenticatorTypePasskey),
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

var nodeAuthenticator = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeAuthenticator,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeAuthenticator),
			"createdAt": entityCreatedAtField(loadAuthenticator),
			"updatedAt": entityUpdatedAtField(loadAuthenticator),
			"expireAfter": &graphql.Field{
				Type: graphql.DateTime,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					info := loadAuthenticator(p.Context, p.Source)
					return info.Map(func(value interface{}) (interface{}, error) {
						p := value.(*authenticator.Info).Password
						if p == nil {
							return nil, nil
						}

						return p.ExpireAfter, nil
					}).Value, nil
				},
			},
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
						claims := value.(*authenticator.Info).ToPublicClaims()
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
		return ctx.Authenticators.Load(id).Value, nil
	},
)

var connAuthenticator = graphqlutil.NewConnectionDef(nodeAuthenticator)

func loadAuthenticator(ctx context.Context, obj interface{}) *graphqlutil.Lazy {
	switch obj := obj.(type) {
	case *authenticator.Info:
		return graphqlutil.NewLazyValue(obj)
	case *authenticator.Ref:
		return GQLContext(ctx).Authenticators.Load(obj.ID)
	default:
		panic(fmt.Sprintf("graphql: unknown authenticator type: %T", obj))
	}
}
