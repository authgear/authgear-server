package graphql

import (
	"fmt"

	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var userSortBy = graphql.NewEnum(graphql.EnumConfig{
	Name: "UserSortBy",
	Values: graphql.EnumValueConfigMap{
		"CREATED_AT": &graphql.EnumValueConfig{
			Value: libuser.SortByCreatedAt,
		},
		"LAST_LOGIN_AT": &graphql.EnumValueConfig{
			Value: libuser.SortByLastLoginAt,
		},
	},
})

var sortDirection = graphql.NewEnum(graphql.EnumConfig{
	Name: "SortDirection",
	Values: graphql.EnumValueConfigMap{
		"ASC": &graphql.EnumValueConfig{
			Value: apimodel.SortDirectionAsc,
		},
		"DESC": &graphql.EnumValueConfig{
			Value: apimodel.SortDirectionDesc,
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
			Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
				"searchKeyword": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"sortBy": &graphql.ArgumentConfig{
					Type: userSortBy,
				},
				"sortDirection": &graphql.ArgumentConfig{
					Type: sortDirection,
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				gqlCtx := GQLContext(p.Context)

				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

				searchKeyword, _ := p.Args["searchKeyword"].(string)

				sortBy, _ := p.Args["sortBy"].(libuser.SortBy)
				sortDirection, _ := p.Args["sortDirection"].(apimodel.SortDirection)

				sortOption := libuser.SortOption{
					SortBy:        sortBy,
					SortDirection: sortDirection,
				}

				var refs []apimodel.PageItemRef
				var result *graphqlutil.PageResult
				var err error
				if searchKeyword == "" {
					refs, result, err = gqlCtx.UserFacade.ListPage(sortOption, pageArgs)
				} else {
					refs, result, err = gqlCtx.UserFacade.SearchPage(searchKeyword, sortOption, pageArgs)
				}
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
		"auditLogs": &graphql.Field{
			Description: "Audit logs",
			Type:        connAuditLog.ConnectionType,
			Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
				"rangeFrom": &graphql.ArgumentConfig{
					Type: graphql.DateTime,
				},
				"rangeTo": &graphql.ArgumentConfig{
					Type: graphql.DateTime,
				},
				"activityTypes": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(auditLogActivityType)),
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// FIXME
				return nil, fmt.Errorf("not yet implemented")
			},
		},
	},
})
