package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node":  nodeDefs.NodeField,
		"nodes": nodeDefs.NodesField,
		"viewer": &graphql.Field{
			Description: "The current viewer",
			Type:        nodeUser,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)

				sessionInfo := session.GetValidSessionInfo(p.Context)
				if sessionInfo == nil {
					return nil, nil
				}

				return ctx.Users.Load(sessionInfo.UserID).Value, nil
			},
		},
		"apps": &graphql.Field{
			Description: "All apps accessible by the viewer",
			Type:        connApp.ConnectionType,
			Args:        relay.ConnectionArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)

				sessionInfo := session.GetValidSessionInfo(p.Context)
				if sessionInfo == nil {
					return nil, nil
				}

				apps, err := ctx.AppService.List(sessionInfo.UserID)
				if err != nil {
					return nil, err
				}
				args := relay.NewConnectionArguments(p.Args)

				out := make([]interface{}, len(apps))
				for i, app := range apps {
					out[i] = app
					ctx.Apps.Prime(app.ID, app)
				}

				return graphqlutil.NewConnectionFromArray(out, args), nil
			},
		},
	},
})
