package graphql

import (
	"errors"
	"fmt"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/web3"
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
			Args: newAnalyticArgs(graphql.FieldConfigArgument{
				"periodical": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(periodicalEnum),
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				periodical := p.Args["periodical"].(string)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = ctx.AuthzService.CheckAccessOfViewer(appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := ctx.AnalyticChartService.GetActiveUserChart(
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
		"totalUserCountChart": &graphql.Field{
			Description: "Total users count chart dataset",
			Type:        totalUserCountChart,
			Args:        newAnalyticArgs(graphql.FieldConfigArgument{}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = ctx.AuthzService.CheckAccessOfViewer(appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := ctx.AnalyticChartService.GetTotalUserCountChart(
					appID,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(chart).Value, nil
			},
		},
		"signupConversionRate": &graphql.Field{
			Description: "Signup conversion rate dashboard data",
			Type:        signupConversionRate,
			Args:        newAnalyticArgs(graphql.FieldConfigArgument{}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = ctx.AuthzService.CheckAccessOfViewer(appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				signupConversionRateData, err := ctx.AnalyticChartService.GetSignupConversionRate(
					appID,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch conversion rate data: %w", err)
				}
				return graphqlutil.NewLazyValue(signupConversionRateData).Value, nil
			},
		},
		"signupByMethodsChart": &graphql.Field{
			Description: "Signup by methods dataset",
			Type:        signupByMethodsChart,
			Args:        newAnalyticArgs(graphql.FieldConfigArgument{}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = ctx.AuthzService.CheckAccessOfViewer(appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := ctx.AnalyticChartService.GetSignupByMethodsChart(
					appID,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(chart).Value, nil
			},
		},
		"subscriptionPlans": &graphql.Field{
			Description: "Available subscription plans",
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(subscriptionPlan))),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				plans, err := ctx.StripeService.FetchSubscriptionPlans()
				if err != nil {
					return nil, err
				}
				return plans, nil
			},
		},
		"nftContractMetadata": &graphql.Field{
			Description: "Fetch NFT Contract Metadata",
			Type:        nftContractMetadata,
			Args: graphql.FieldConfigArgument{
				"appID": &graphql.ArgumentConfig{
					Type: graphql.ID,
				},
				"contractID": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				appNodeID := p.Args["appID"].(string)
				contractURL := p.Args["contractID"].(string)

				contractID, err := web3.ParseContractID(contractURL)
				if err != nil {
					return nil, err
				}

				metadata, err := ctx.NFTService.GetContractMetadata(appNodeID, *contractID)
				if err != nil {
					return nil, err
				}

				return metadata, nil
			},
		},
	},
})
