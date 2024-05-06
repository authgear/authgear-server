package graphql

import (
	"context"
	"errors"
	"time"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

var oauthSSOProviderClientSecret = graphql.NewObject(graphql.ObjectConfig{
	Name:        "OAuthSSOProviderClientSecret",
	Description: "OAuth client secret",
	Fields: graphql.Fields{
		"alias": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"clientSecret": &graphql.Field{
			Type: graphql.String,
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

var oauthClientSecretKey = graphql.NewObject(graphql.ObjectConfig{
	Name:        "oauthClientSecretKey",
	Description: "OAuth client secret key",
	Fields: graphql.Fields{
		"keyID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"key": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var oauthClientSecretItem = graphql.NewObject(graphql.ObjectConfig{
	Name:        "oauthClientSecretItem",
	Description: "OAuth client secret item",
	Fields: graphql.Fields{
		"clientID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"keys": &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(oauthClientSecretKey)),
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

type AppSecretKey string

const (
	AppSecretKeyOauthSSOProviderClientSecrets AppSecretKey = "oauthSSOProviderClientSecrets" // nolint:gosec
	AppSecretKeyWebhookSecret                 AppSecretKey = "webhookSecret"
	AppSecretKeyAdminAPISecrets               AppSecretKey = "adminAPISecrets"
	AppSecretKeySmtpSecret                    AppSecretKey = "smtpSecret"
	AppSecretKeyOauthClientSecrets            AppSecretKey = "oauthClientSecrets" // nolint:gosec
)

var secretConfig = graphql.NewObject(graphql.ObjectConfig{
	Name:        "SecretConfig",
	Description: "The content of authgear.secrets.yaml",
	Fields: graphql.Fields{
		string(AppSecretKeyOauthSSOProviderClientSecrets): &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(oauthSSOProviderClientSecret)),
		},
		string(AppSecretKeyWebhookSecret): &graphql.Field{
			Type: webhookSecret,
		},
		string(AppSecretKeyAdminAPISecrets): &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(adminAPISecret)),
		},
		string(AppSecretKeySmtpSecret): &graphql.Field{
			Type: smtpSecret,
		},
		string(AppSecretKeyOauthClientSecrets): &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(oauthClientSecretItem)),
		},
	},
})

var appSecretKey = graphql.NewEnum(graphql.EnumConfig{
	Name: "AppSecretKey",
	Values: graphql.EnumValueConfigMap{
		"OAUTH_SSO_PROVIDER_CLIENT_SECRETS": &graphql.EnumValueConfig{
			Value: AppSecretKeyOauthSSOProviderClientSecrets,
		},
		"WEBHOOK_SECRET": &graphql.EnumValueConfig{
			Value: AppSecretKeyWebhookSecret,
		},
		"ADMIN_API_SECRETS": &graphql.EnumValueConfig{
			Value: AppSecretKeyAdminAPISecrets,
		},
		"SMTP_SECRET": &graphql.EnumValueConfig{
			Value: AppSecretKeySmtpSecret,
		},
		"OAUTH_CLIENT_SECRETS": &graphql.EnumValueConfig{
			Value: AppSecretKeyOauthClientSecrets,
		},
	},
})

var secretKeyToConfigKeyMap map[AppSecretKey]config.SecretKey = map[AppSecretKey]config.SecretKey{
	AppSecretKeyOauthSSOProviderClientSecrets: config.OAuthSSOProviderCredentialsKey,
	AppSecretKeyWebhookSecret:                 config.WebhookKeyMaterialsKey,
	AppSecretKeyAdminAPISecrets:               config.AdminAPIAuthKeyKey,
	AppSecretKeySmtpSecret:                    config.SMTPServerCredentialsKey,
	AppSecretKeyOauthClientSecrets:            config.OAuthClientCredentialsKey,
}

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
						return []*model.AppResource{}, nil
					}
					descriptedPaths, err := appResMgr.AssociateDescriptor(paths...)
					if err != nil {
						return nil, err
					}

					var appRes []*model.AppResource
					for _, p := range descriptedPaths {
						appRes = append(appRes, &model.AppResource{
							Context:        app.Context,
							DescriptedPath: resource.DescriptedPath(p),
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
					config, _, err := ctx.AppService.LoadRawAppConfig(app)
					return config, err
				},
			},
			"rawAppConfigChecksum": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					app := p.Source.(*model.App)
					_, checksum, err := ctx.AppService.LoadRawAppConfig(app)
					return checksum, err
				},
			},
			"secretConfig": &graphql.Field{
				Type: graphql.NewNonNull(secretConfig),
				Args: graphql.FieldConfigArgument{
					"token": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					var token string
					if tokenRaw, ok := p.Args["token"]; ok {
						if t, ok := tokenRaw.(string); ok {
							token = t
						}
					}
					app := p.Source.(*model.App)
					sessionInfo := session.GetValidSessionInfo(p.Context)
					if sessionInfo == nil {
						return nil, Unauthenticated.New("only authenticated users can view app secret")
					}

					secretConfig, _, err := ctx.AppService.LoadAppSecretConfig(app, sessionInfo, token)
					if err != nil {
						return nil, err
					}

					return secretConfig, nil
				},
			},
			"secretConfigChecksum": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					var token string
					if tokenRaw, ok := p.Args["token"]; ok {
						if t, ok := tokenRaw.(string); ok {
							token = t
						}
					}
					app := p.Source.(*model.App)
					sessionInfo := session.GetValidSessionInfo(p.Context)
					if sessionInfo == nil {
						return nil, Unauthenticated.New("only authenticated users can view app secret checksum")
					}

					_, checksum, err := ctx.AppService.LoadAppSecretConfig(app, sessionInfo, token)
					if err != nil {
						return nil, err
					}
					return checksum, nil
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
					ctx := GQLContext(p.Context)
					appID := p.Source.(*model.App).ID
					planName := p.Source.(*model.App).Context.PlanName
					date := p.Args["date"].(time.Time)

					plans, err := ctx.StripeService.FetchSubscriptionPlans()
					if err != nil {
						return nil, err
					}

					subscriptionUsage, err := ctx.SubscriptionService.GetSubscriptionUsage(appID, planName, date, plans)
					if err != nil {
						return nil, err
					}

					return subscriptionUsage, nil
				},
			},
			"subscription": &graphql.Field{
				Type: subscription,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					appID := p.Source.(*model.App).ID

					subscription, err := ctx.SubscriptionService.GetSubscription(appID)
					if errors.Is(err, service.ErrSubscriptionNotFound) {
						return nil, nil
					} else if err != nil {
						return nil, err
					}

					return subscription, nil
				},
			},
			"isProcessingSubscription": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					appID := p.Source.(*model.App).ID
					customerID, err := ctx.SubscriptionService.GetLastProcessingCustomerID(appID)
					if err != nil {
						return nil, err
					}

					return customerID != nil, nil
				},
			},
			"lastStripeError": &graphql.Field{
				Type: StripeError,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					appID := p.Source.(*model.App).ID
					customerID, err := ctx.SubscriptionService.GetLastProcessingCustomerID(appID)
					if err != nil {
						return nil, err
					}
					if customerID == nil {
						return nil, err
					}

					stripeError, err := ctx.StripeService.GetLastPaymentError(*customerID)
					if err != nil {
						return nil, err
					}

					return stripeError, nil
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
			"viewer": &graphql.Field{
				Type: graphql.NewNonNull(collaborator),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)

					sessionInfo := session.GetValidSessionInfo(p.Context)
					if sessionInfo == nil {
						return nil, Unauthenticated.New("only authenticated users can query app viewer")
					}

					app := p.Source.(*model.App)
					collaborator, err := ctx.CollaboratorService.GetCollaboratorByAppAndUser(app.ID, sessionInfo.UserID)
					if err != nil {
						return nil, err
					}

					return ctx.Collaborators.Load(collaborator.ID).Value, nil
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
			"nftCollections": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nftCollection))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := GQLContext(p.Context)
					config := p.Source.(*model.App).Context.Config.AppConfig
					if config.Web3 == nil {
						return []interface{}{}, nil
					}

					if config.Web3.NFT == nil {
						return []interface{}{}, nil
					}

					if len(config.Web3.NFT.Collections) == 0 {
						return []interface{}{}, nil
					}

					contractIDs := make([]web3.ContractID, 0, len(config.Web3.NFT.Collections))
					for _, url := range config.Web3.NFT.Collections {
						contractID, err := web3.ParseContractID(url)
						if err != nil {
							return nil, err
						}

						contractIDs = append(contractIDs, *contractID)
					}

					collections, err := ctx.NFTService.GetContractMetadata(contractIDs)
					if err != nil {
						return nil, err
					}

					return collections, nil
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
