package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var checkCollaboratorInvitationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CheckCollaboratorInvitationPayload",
	Fields: graphql.Fields{
		"isCodeValid": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"isInvitee": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"appID": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

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

				// Access control is not needed here cause List returns accessible apps.
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
		"checkCollaboratorInvitation": &graphql.Field{
			Description: "Check whether the viewer can accept the collaboration invitation",
			Type:        graphql.NewNonNull(checkCollaboratorInvitationPayload),
			Args:        graphql.FieldConfigArgument{"code": &graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)}},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)

				code := p.Args["code"].(string)

				invitation, err := ctx.CollaboratorService.GetInvitationWithCode(code)
				if err != nil {
					return graphqlutil.NewLazyValue(map[string]interface{}{
						"isCodeValid": false,
						"isInvitee": false,
						"appID": "",
					}).Value, nil
				}

				sessionInfo := session.GetValidSessionInfo(p.Context)
				if sessionInfo == nil {
					return graphqlutil.NewLazyValue(map[string]interface{}{
						"isCodeValid": true,
						"isInvitee": false,
						"appID": invitation.AppID,
					}).Value, nil
				}
				actorID := sessionInfo.UserID

				err = ctx.CollaboratorService.CheckInviteeEmail(invitation, actorID)
				if err != nil {
					return graphqlutil.NewLazyValue(map[string]interface{}{
						"isCodeValid": true,
						"isInvitee": false,
						"appID": invitation.AppID,
					}).Value, nil
				}

				return graphqlutil.NewLazyValue(map[string]interface{}{
						"isCodeValid": true,
						"isInvitee": true,
						"appID": invitation.AppID,
				}).Value, nil
			},
		},
	},
})
