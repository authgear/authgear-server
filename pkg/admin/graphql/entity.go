package graphql

import (
	"context"
	"fmt"
	"reflect"

	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var nodeDefs = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
		// If the ID is invalid, we should return null instead of returning an error.
		// This behavior conforms the schema.
		resolvedID := relay.FromGlobalID(id)
		if resolvedID == nil {
			return nil, nil
		}
		resolver, ok := resolvers[resolvedID.Type]
		if !ok {
			return nil, nil
		}
		return resolver(GQLContext(ctx), resolvedID.ID)
	},
	TypeResolve: func(params graphql.ResolveTypeParams) *graphql.Object {
		objType, ok := entityTypes[reflect.TypeOf(params.Value)]
		if !ok {
			panic(fmt.Sprintf("graphql: unknown value type: %T", params.Value))
		}
		return objType
	},
})

var entityInterface = graphql.NewInterface(graphql.InterfaceConfig{
	Name: "Entity",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of entity",
		},
		"createdAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.DateTime),
			Description: "The creation time of entity",
		},
		"updatedAt": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.DateTime),
			Description: "The update time of entity",
		},
	},
	ResolveType: func(params graphql.ResolveTypeParams) *graphql.Object {
		objType, ok := entityTypes[reflect.TypeOf(params.Value)]
		if !ok {
			panic(fmt.Sprintf("graphql: unknown value type: %T", params.Value))
		}
		return objType
	},
})

type EntityResolver func(ctx *Context, id string) (interface{}, error)

var resolvers = map[string]EntityResolver{}
var entityTypes = map[reflect.Type]*graphql.Object{}

func entity(schema *graphql.Object, modelType interface{}, resolver EntityResolver) *graphql.Object {
	resolvers[schema.Name()] = resolver
	entityTypes[reflect.TypeOf(modelType)] = schema
	return schema
}

type EntityRef interface {
	GetMeta() model.Meta
}

func entityIDField(typeName string, idFn func(obj interface{}) (string, error)) *graphql.Field {
	return relay.GlobalIDField(
		typeName,
		func(obj interface{}, info graphql.ResolveInfo, ctx context.Context) (string, error) {
			if idFn != nil {
				return idFn(obj)
			}
			meta := obj.(EntityRef).GetMeta()
			return meta.ID, nil
		},
	)
}

func entityCreatedAtField(objFn func(ctx context.Context, obj interface{}) *graphqlutil.Lazy) *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewNonNull(graphql.DateTime),
		Description: "The creation time of entity",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			obj := graphqlutil.NewLazyValue(p.Source)
			if objFn != nil {
				obj = objFn(p.Context, p.Source)
			}
			obj = obj.Map(func(value interface{}) (interface{}, error) {
				meta := value.(EntityRef).GetMeta()
				return meta.CreatedAt, nil
			})
			return obj.Value, nil
		},
	}
}

func entityUpdatedAtField(objFn func(ctx context.Context, obj interface{}) *graphqlutil.Lazy) *graphql.Field {
	return &graphql.Field{
		Type:        graphql.NewNonNull(graphql.DateTime),
		Description: "The update time of entity",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			obj := graphqlutil.NewLazyValue(p.Source)
			if objFn != nil {
				obj = objFn(p.Context, p.Source)
			}
			obj = obj.Map(func(value interface{}) (interface{}, error) {
				meta := value.(EntityRef).GetMeta()
				return meta.UpdatedAt, nil
			})
			return obj.Value, nil
		},
	}
}
