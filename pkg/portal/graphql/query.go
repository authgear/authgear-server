package graphql

import (
	"errors"
	"fmt"
	"time"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var checkCollaboratorInvitationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CheckCollaboratorInvitationPayload",
	Fields: graphql.Fields{
		"isInvitee": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"appID":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
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
			Type:        checkCollaboratorInvitationPayload,
			Args:        graphql.FieldConfigArgument{"code": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)

				code := p.Args["code"].(string)

				invitation, err := ctx.CollaboratorService.GetInvitationWithCode(code)
				if err != nil {
					if errors.Is(err, service.ErrCollaboratorInvitationInvalidCode) {
						return nil, nil
					}
					return nil, err
				}

				sessionInfo := session.GetValidSessionInfo(p.Context)
				if sessionInfo == nil {
					return graphqlutil.NewLazyValue(map[string]interface{}{
						"isInvitee": false,
						"appID":     invitation.AppID,
					}).Value, nil
				}
				actorID := sessionInfo.UserID

				err = ctx.CollaboratorService.CheckInviteeEmail(invitation, actorID)
				if err != nil {
					if errors.Is(err, service.ErrCollaboratorInvitationInvalidEmail) {
						return graphqlutil.NewLazyValue(map[string]interface{}{
							"isInvitee": false,
							"appID":     invitation.AppID,
						}).Value, nil
					}
					return nil, err
				}

				return graphqlutil.NewLazyValue(map[string]interface{}{
					"isInvitee": true,
					"appID":     invitation.AppID,
				}).Value, nil
			},
		},
		"activeUserChart": &graphql.Field{
			Description: "Active users chart dataset",
			Type:        activeUserChart,
			Args: graphql.FieldConfigArgument{
				"appID": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Target app ID.",
				},
				"periodical": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(periodicalEnum),
				},
				"rangeFrom": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphqlutil.Date),
				},
				"rangeTo": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphqlutil.Date),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				appID := p.Args["appID"].(string)
				periodical := p.Args["periodical"].(string)

				ctx := GQLContext(p.Context)
				var rangeFrom *time.Time
				if t, ok := p.Args["rangeFrom"].(time.Time); ok {
					rangeFrom = &t
				}

				var rangeTo *time.Time
				if t, ok := p.Args["rangeTo"].(time.Time); ok {
					rangeTo = &t
				}

				err := checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := ctx.AnalyticChartService.GetActiveUserChat(
					appID,
					periodical,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(chart).Value, nil
			},
		},
		"signupSummary": &graphql.Field{
			Description: "Signup summary for analytic dashboard",
			Type:        signupSummary,
			Args: graphql.FieldConfigArgument{
				"appID": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Target app ID.",
				},
				"rangeFrom": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphqlutil.Date),
				},
				"rangeTo": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphqlutil.Date),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				appID := p.Args["appID"].(string)
				var rangeFrom *time.Time
				if t, ok := p.Args["rangeFrom"].(time.Time); ok {
					rangeFrom = &t
				}

				var rangeTo *time.Time
				if t, ok := p.Args["rangeTo"].(time.Time); ok {
					rangeTo = &t
				}

				err := checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				signupSummary, err := ctx.AnalyticChartService.GetSignupSummary(appID, *rangeFrom, *rangeTo)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(signupSummary).Value, nil
			},
		},
	},
})
