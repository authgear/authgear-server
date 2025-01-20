package graphql

import (
	"context"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

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
		objType, ok := typeMapping[reflect.TypeOf(params.Value)]
		if !ok {
			panic(fmt.Sprintf("graphql: unknown value type: %T", params.Value))
		}
		return objType
	},
})

type EntityRef interface {
	GetMeta() model.Meta
}

func entityIDField(typeName string) *graphql.Field {
	return relay.GlobalIDField(
		typeName,
		func(obj interface{}, info graphql.ResolveInfo, ctx context.Context) (string, error) {
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
