package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var searchUsersSortBy = graphql.NewEnum(graphql.EnumConfig{
	Name: "SearchUsersSortBy",
	Values: graphql.EnumValueConfigMap{
		"CREATED_AT": &graphql.EnumValueConfig{
			Value: libes.QueryUserSortByCreatedAt,
		},
		"LAST_LOGIN_AT": &graphql.EnumValueConfig{
			Value: libes.QueryUserSortByLastLoginAt,
		},
	},
})

var sortDirection = graphql.NewEnum(graphql.EnumConfig{
	Name: "SortDirection",
	Values: graphql.EnumValueConfigMap{
		"ASC": &graphql.EnumValueConfig{
			Value: libes.SortDirectionAsc,
		},
		"DESC": &graphql.EnumValueConfig{
			Value: libes.SortDirectionDesc,
		},
	},
})

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node":  nodeDefs.NodeField,
		"nodes": nodeDefs.NodesField,
		"users": &graphql.Field{
			Description: "All users",
			Type:        connUser.ConnectionType,
			Args:        relay.ConnectionArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				args := relay.NewConnectionArguments(p.Args)
				gqlCtx := GQLContext(p.Context)
				refs, result, err := gqlCtx.UserFacade.QueryPage(graphqlutil.NewPageArgs(args))
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.Users.Load(ref.ID),
						Cursor: graphqlutil.Cursor(ref.Cursor),
					})
				}

				return graphqlutil.NewConnectionFromResult(lazyItems, result)
			},
		},
		"searchUsers": &graphql.Field{
			Description: "Search users",
			Type:        connUser.ConnectionType,
			Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
				"searchKeyword": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"sortBy": &graphql.ArgumentConfig{
					Type: searchUsersSortBy,
				},
				"sortDirection": &graphql.ArgumentConfig{
					Type: sortDirection,
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				gqlCtx := GQLContext(p.Context)
				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))
				searchKeyword, _ := p.Args["searchKeyword"].(string)
				sortBy, _ := p.Args["sortBy"].(libes.QueryUserSortBy)
				sortDirection, _ := p.Args["sortDirection"].(libes.SortDirection)

				opts := &libes.QueryUserOptions{
					SearchKeyword: searchKeyword,
					First:         *pageArgs.First,
					After:         apimodel.PageCursor(pageArgs.After),
					SortBy:        sortBy,
					SortDirection: sortDirection,
				}

				refs, result, err := gqlCtx.UserFacade.SearchPage(pageArgs, opts)
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.Users.Load(ref.ID),
						Cursor: graphqlutil.Cursor(ref.Cursor),
					})
				}

				return graphqlutil.NewConnectionFromResult(lazyItems, result)
			},
		},
	},
})
