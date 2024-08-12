package graphql

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeIdentity = "Identity"

var identityType = graphql.NewEnum(graphql.EnumConfig{
	Name: "IdentityType",
	Values: graphql.EnumValueConfigMap{
		"LOGIN_ID": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypeLoginID),
		},
		"OAUTH": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypeOAuth),
		},
		"ANONYMOUS": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypeAnonymous),
		},
		"BIOMETRIC": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypeBiometric),
		},
		"PASSKEY": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypePasskey),
		},
		"SIWE": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypeSIWE),
		},
		"LDAP": &graphql.EnumValueConfig{
			Value: string(model.IdentityTypeLDAP),
		},
	},
})

var nodeIdentity = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeIdentity,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeIdentity),
			"createdAt": entityCreatedAtField(loadIdentity),
			"updatedAt": entityUpdatedAtField(loadIdentity),
			"type": &graphql.Field{
				Type: graphql.NewNonNull(identityType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ref := p.Source.(interface{ ToRef() *model.IdentityRef }).ToRef()
					return string(ref.Type), nil
				},
			},
			"claims": &graphql.Field{
				Type: graphql.NewNonNull(IdentityClaims),
				Args: map[string]*graphql.ArgumentConfig{
					"names": {Type: graphql.NewList(graphql.NewNonNull(graphql.String))},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					names, hasNames := p.Args["names"].([]interface{})
					info := loadIdentity(p.Context, p.Source)
					claims := info.Map(func(value interface{}) (interface{}, error) {
						info := value.(*identity.Info)
						m := info.ToModel()
						claims := m.Claims
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
		return ctx.Identities.Load(id).Value, nil
	},
)

var connIdentity = graphqlutil.NewConnectionDef(nodeIdentity)

func loadIdentity(ctx context.Context, obj interface{}) *graphqlutil.Lazy {
	switch obj := obj.(type) {
	case *identity.Info:
		return graphqlutil.NewLazyValue(obj)
	case *model.IdentityRef:
		return GQLContext(ctx).Identities.Load(obj.ID)
	default:
		panic(fmt.Sprintf("graphql: unknown identity type: %T", obj))
	}
}
