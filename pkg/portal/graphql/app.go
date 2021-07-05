package graphql

import (
	"context"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/util/resources"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeApp = "App"

var nodeApp = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeApp,
		Description: "Authgear app",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeApp, func(obj interface{}, info graphql.ResolveInfo, ctx context.Context) (string, error) {
				return obj.(*model.App).ID, nil
			}),
			"resources": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(appResource))),
				Args: graphql.FieldConfigArgument{
					"paths": &graphql.ArgumentConfig{
						Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var paths []string
					if argPaths, ok := p.Args["paths"]; ok {
						for _, path := range argPaths.([]interface{}) {
							paths = append(paths, path.(string))
						}
					}

					app := p.Source.(*model.App)
					if len(paths) == 0 {
						var err error
						paths, err = resources.List(app.Context.Resources)
						if err != nil {
							return nil, err
						}
					}
					descriptedPaths, err := resources.AssociateDescriptor(app.Context.Resources, paths...)
					if err != nil {
						return nil, err
					}

					var appRes []*model.AppResource
					for _, p := range descriptedPaths {
						appRes = append(appRes, &model.AppResource{
							Context:        app.Context,
							DescriptedPath: p,
						})
					}
					return appRes, nil
				},
			},
			"rawAppConfig": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					return ctx.AppService.LoadRawAppConfig(app)
				},
			},
			"rawSecretConfig": &graphql.Field{
				Type: graphql.NewNonNull(SecretConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					return ctx.AppService.LoadRawSecretConfig(app)
				},
			},
			"effectiveAppConfig": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*model.App).Context.Config.AppConfig, nil
				},
			},
			"effectiveFeatureConfig": &graphql.Field{
				Type: graphql.NewNonNull(FeatureConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*model.App).Context.Config.FeatureConfig, nil
				},
			},
			"planName": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*model.App).Context.PlanName, nil
				},
			},
			"domains": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(domain))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					domains, err := ctx.DomainService.ListDomains(app.ID)
					if err != nil {
						return nil, err
					}

					ids := make([]interface{}, len(domains))
					for i, domain := range domains {
						id := domain.ID
						ids[i] = id
						ctx.Domains.Prime(id, domain)
					}

					return ctx.Domains.LoadMany(ids).Value, nil
				},
			},
			"collaborators": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(collaborator))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					collaborators, err := ctx.CollaboratorService.ListCollaborators(app.ID)
					if err != nil {
						return nil, err
					}

					ids := make([]interface{}, len(collaborators))
					for i, collaborator := range collaborators {
						id := collaborator.ID
						ids[i] = id
						ctx.Collaborators.Prime(id, collaborator)
					}

					return ctx.Collaborators.LoadMany(ids).Value, nil
				},
			},
			"collaboratorInvitations": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(collaboratorInvitation))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					invitations, err := ctx.CollaboratorService.ListInvitations(app.ID)
					if err != nil {
						return nil, err
					}

					ids := make([]interface{}, len(invitations))
					for i, invitation := range invitations {
						id := invitation.ID
						ids[i] = id
						ctx.CollaboratorInvitations.Prime(id, invitation)
					}

					return ctx.CollaboratorInvitations.LoadMany(ids).Value, nil
				},
			},
		},
	}),
	&model.App{},
	func(ctx context.Context, id string) (interface{}, error) {
		gqlCtx := GQLContext(ctx)
		// return nil without error for both inaccessible / not found apps
		_, err := gqlCtx.AuthzService.CheckAccessOfViewer(id)
		if err != nil {
			return nil, nil
		}
		return gqlCtx.Apps.Load(id).Value, nil
	},
)

var connApp = graphqlutil.NewConnectionDef(nodeApp)
