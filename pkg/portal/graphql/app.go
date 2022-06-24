package graphql

import (
	"context"
	"errors"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var oauthClientSecret = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OAuthClientSecret",
	Description: "OAuth client secret",
	Fields: graphql.Fields{
		"alias": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"clientSecret": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var webhookSecret = graphql.NewObject(graphql.ObjectConfig{
	Name:        "WebhookSecret",
	Description: "Webhook secret",
	Fields: graphql.Fields{
		"secret": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var adminAPISecret = graphql.NewObject(graphql.ObjectConfig{
	Name:        "AdminAPISecret",
	Description: "Admin API secret",
	Fields: graphql.Fields{
		"keyID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"publicKeyPEM": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"privateKeyPEM": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var smtpSecret = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SMTPSecret",
	Description: "SMTP secret",
	Fields: graphql.Fields{
		"host": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"port": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"username": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"password": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var secretConfig = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SecretConfig",
	Description: "The content of authgear.secrets.yaml",
	Fields: graphql.Fields{
		"oauthClientSecrets": &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(oauthClientSecret)),
		},
		"webhookSecret": &graphql.Field{
			Type: webhookSecret,
		},
		"adminAPISecrets": &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(adminAPISecret)),
		},
		"smtpSecret": &graphql.Field{
			Type: smtpSecret,
		},
	},
})

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
							path := path.(string)
							if path == configsource.AuthgearYAML {
								return nil, errors.New("direct access on authgear.yaml is disallowed")
							}
							if path == configsource.AuthgearSecretYAML {
								return nil, errors.New("direct access on authgear.secrets.yaml is disallowed")
							}
							paths = append(paths, path)
						}
					}

					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					appResMgr := ctx.AppResMgrFactory.NewManagerWithAppContext(app.Context)
					if len(paths) == 0 {
						var err error
						paths, err = appResMgr.List()
						if err != nil {
							return nil, err
						}
					}
					descriptedPaths, err := appResMgr.AssociateDescriptor(paths...)
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
			"secretConfig": &graphql.Field{
				Type: graphql.NewNonNull(secretConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					sessionInfo := session.GetValidSessionInfo(p.Context)
					secretConfig, err := ctx.AppService.LoadAppSecretConfig(app, sessionInfo)
					if err != nil {
						return nil, err
					}
					return secretConfig, nil
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
			"subscriptionUsage": &graphql.Field{
				Type: subscriptionUsage,
				Args: graphql.FieldConfigArgument{
					"date": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.DateTime),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
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
			"tutorialStatus": &graphql.Field{
				Type: graphql.NewNonNull(tutorialStatus),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)

					entry, err := ctx.TutorialService.Get(app.ID)
					if err != nil {
						return nil, err
					}

					return entry, nil
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
