package graphql

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
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
				"groupKeys": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
				},
				"roleKeys": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
				},
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
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

				searchKeyword, _ := p.Args["searchKeyword"].(string)

				sortBy, _ := p.Args["sortBy"].(libuser.SortBy)
				sortDirection, _ := p.Args["sortDirection"].(apimodel.SortDirection)

				sortOption := libuser.SortOption{
					SortBy:        sortBy,
					SortDirection: sortDirection,
				}

				groupKeyIfaces, _ := p.Args["groupKeys"].([]interface{})
				groupKeys := make([]string, len(groupKeyIfaces))
				for i := range groupKeyIfaces {
					groupKeys[i] = groupKeyIfaces[i].(string)
				}

				roleKeyIfaces, _ := p.Args["roleKeys"].([]interface{})
				roleKeys := make([]string, len(roleKeyIfaces))
				for i := range roleKeyIfaces {
					roleKeys[i] = roleKeyIfaces[i].(string)
				}

				listOption := libuser.ListOptions{
					SortOption: sortOption,
				}

				filterOptions := libuser.FilterOptions{
					RoleKeys:  roleKeys,
					GroupKeys: groupKeys,
				}

				var refs []apimodel.PageItemRef
				var result *graphqlutil.PageResult
				var err error
				if searchKeyword == "" && !filterOptions.IsFilterEnabled() {
					refs, result, err = gqlCtx.UserFacade.ListPage(ctx, listOption, pageArgs)
				} else {
					refs, result, err = gqlCtx.UserFacade.SearchPage(ctx, searchKeyword, filterOptions, sortOption, pageArgs)
				}
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.Users.Load(ctx, ref.ID),
						Cursor: graphqlutil.Cursor(ref.Cursor),
					})
				}

				return graphqlutil.NewConnectionFromResult(lazyItems, result)
			},
		},
		"roles": &graphql.Field{
			Description: "All roles",
			Type:        connRole.ConnectionType,
			Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
				"searchKeyword": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"excludedIDs": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.ID)),
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

				searchKeyword, _ := p.Args["searchKeyword"].(string)

				excludedIDsIfaces, _ := p.Args["excludedIDs"].([]interface{})
				excludedNodeIDs := make([]string, len(excludedIDsIfaces))
				for i, v := range excludedIDsIfaces {
					excludedNodeIDs[i] = v.(string)
				}

				excludedIDs := make([]string, len(excludedIDsIfaces))
				for i, v := range excludedNodeIDs {
					resolvedNodeID := relay.FromGlobalID(v)
					if resolvedNodeID == nil || resolvedNodeID.Type != typeRole {
						return nil, apierrors.NewInvalid("invalid role ID")
					}
					excludedIDs[i] = resolvedNodeID.ID
				}

				options := &rolesgroups.ListRolesOptions{
					SearchKeyword: searchKeyword,
					ExcludedIDs:   excludedIDs,
				}

				refs, result, err := gqlCtx.RolesGroupsFacade.ListRoles(ctx, options, pageArgs)
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.Roles.Load(ctx, ref.ID),
						Cursor: graphqlutil.Cursor(ref.Cursor),
					})
				}

				return graphqlutil.NewConnectionFromResult(lazyItems, result)
			},
		},
		"groups": &graphql.Field{
			Description: "All groups",
			Type:        connGroup.ConnectionType,
			Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
				"searchKeyword": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"excludedIDs": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.ID)),
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

				searchKeyword, _ := p.Args["searchKeyword"].(string)

				excludedIDsIfaces, _ := p.Args["excludedIDs"].([]interface{})
				excludedNodeIDs := make([]string, len(excludedIDsIfaces))
				for i, v := range excludedIDsIfaces {
					excludedNodeIDs[i] = v.(string)
				}

				excludedIDs := make([]string, len(excludedIDsIfaces))
				for i, v := range excludedNodeIDs {
					resolvedNodeID := relay.FromGlobalID(v)
					if resolvedNodeID == nil || resolvedNodeID.Type != typeGroup {
						return nil, apierrors.NewInvalid("invalid group ID")
					}
					excludedIDs[i] = resolvedNodeID.ID
				}

				options := &rolesgroups.ListGroupsOptions{
					SearchKeyword: searchKeyword,
					ExcludedIDs:   excludedIDs,
				}

				refs, result, err := gqlCtx.RolesGroupsFacade.ListGroups(ctx, options, pageArgs)
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.Groups.Load(ctx, ref.ID),
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
				"userIDs": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.ID)),
				},
				"emailAddresses": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
				},
				"phoneNumbers": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
				},
				"sortDirection": &graphql.ArgumentConfig{
					Type: sortDirection,
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

				sortDirection, _ := p.Args["sortDirection"].(apimodel.SortDirection)

				var rangeFrom *time.Time
				if t, ok := p.Args["rangeFrom"].(time.Time); ok {
					rangeFrom = &t
				}

				var rangeTo *time.Time
				if t, ok := p.Args["rangeTo"].(time.Time); ok {
					rangeTo = &t
				}

				var activityTypes []string
				if arr, ok := p.Args["activityTypes"].([]interface{}); ok {
					for _, v := range arr {
						if s, ok := v.(string); ok {
							activityTypes = append(activityTypes, s)
						}
					}
				}

				var userIDs []string
				if arr, ok := p.Args["userIDs"].([]interface{}); ok {
					for _, v := range arr {
						if s, ok := v.(string); ok {
							resolvedNodeID := relay.FromGlobalID(s)
							if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
								return nil, apierrors.NewInvalid("invalid user IDs")
							}
							userIDs = append(userIDs, resolvedNodeID.ID)
						}
					}
				}

				var emailAddresses []string
				if arr, ok := p.Args["emailAddresses"].([]interface{}); ok {
					for _, v := range arr {
						if s, ok := v.(string); ok {
							emailAddresses = append(emailAddresses, s)
						}
					}
				}

				var phoneNumbers []string
				if arr, ok := p.Args["phoneNumbers"].([]interface{}); ok {
					for _, v := range arr {
						if s, ok := v.(string); ok {
							phoneNumbers = append(phoneNumbers, s)
						}
					}
				}

				queryOptions := audit.QueryPageOptions{
					RangeFrom:      rangeFrom,
					RangeTo:        rangeTo,
					ActivityTypes:  activityTypes,
					SortDirection:  sortDirection,
					UserIDs:        userIDs,
					EmailAddresses: emailAddresses,
					PhoneNumbers:   phoneNumbers,
				}

				refs, result, err := gqlCtx.AuditLogFacade.QueryPage(ctx, queryOptions, pageArgs)
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.AuditLogs.Load(ctx, ref.ID),
						Cursor: graphqlutil.Cursor(ref.Cursor),
					})
				}

				return graphqlutil.NewConnectionFromResult(lazyItems, result)
			},
		},
		"resources": &graphql.Field{
			Description: "All resources",
			Type:        connResource.ConnectionType,
			Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
				"searchKeyword": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"clientID": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

				searchKeyword, _ := p.Args["searchKeyword"].(string)
				clientID, _ := p.Args["clientID"].(string)

				options := &resourcescope.ListResourcesOptions{
					SearchKeyword: searchKeyword,
					ClientID:      clientID,
				}

				refs, result, err := gqlCtx.ResourceScopeFacade.ListResources(ctx, options, pageArgs)
				if err != nil {
					return nil, err
				}

				var lazyItems []graphqlutil.LazyItem
				for _, ref := range refs {
					lazyItems = append(lazyItems, graphqlutil.LazyItem{
						Lazy:   gqlCtx.Resources.Load(ctx, ref.ID),
						Cursor: graphqlutil.Cursor(ref.Cursor),
					})
				}

				return graphqlutil.NewConnectionFromResult(lazyItems, result)
			},
		},
		"getUsersByStandardAttribute": &graphql.Field{
			Description: "Get users by standardAttribute, attributeName must be email, phone_number or preferred_username.",
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeUser))),
			Args: graphql.FieldConfigArgument{
				"attributeName": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"attributeValue": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				attributeName, _ := p.Args["attributeName"].(string)
				attributeValue, _ := p.Args["attributeValue"].(string)

				userIDs, err := gqlCtx.UserFacade.GetUsersByStandardAttribute(ctx, attributeName, attributeValue)
				if err != nil {
					return nil, err
				}

				return slice.Map(userIDs, func(userID string) interface{} {
					lazyItem, _ := graphqlutil.NewLazyValue(gqlCtx.Users.Load(ctx, userID)).Value()
					return lazyItem
				}), err
			},
		},
		"getUserByLoginID": &graphql.Field{
			Description: "Get user by Login ID.",
			Type:        nodeUser,
			Args: graphql.FieldConfigArgument{
				"loginIDKey": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"loginIDValue": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				loginIDKey, _ := p.Args["loginIDKey"].(string)
				loginIDValue, _ := p.Args["loginIDValue"].(string)

				userID, err := gqlCtx.UserFacade.GetUserByLoginID(ctx, loginIDKey, loginIDValue)
				if errors.Is(err, api.ErrUserNotFound) {
					// For user not found error, just return nil instead of return error
					return nil, nil
				} else if err != nil {
					return nil, err
				}

				return gqlCtx.Users.Load(ctx, userID).Value()
			},
		},
		"getUserByOAuth": &graphql.Field{
			Description: "Get user by OAuth Alias and user ID.",
			Type:        nodeUser,
			Args: graphql.FieldConfigArgument{
				"oauthProviderAlias": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"oauthProviderUserID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				oauthProviderAlias, _ := p.Args["oauthProviderAlias"].(string)
				oauthProviderUserID, _ := p.Args["oauthProviderUserID"].(string)

				userID, err := gqlCtx.UserFacade.GetUserByOAuth(ctx, oauthProviderAlias, oauthProviderUserID)
				if errors.Is(err, api.ErrUserNotFound) {
					// For user not found error, just return nil instead of return error
					return nil, nil
				} else if err != nil {
					return nil, err
				}

				return gqlCtx.Users.Load(ctx, userID).Value()
			},
		},
	},
})
