package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/portal/model"
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
			"rawConfigFile": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Args: graphql.FieldConfigArgument{
					"path": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					path := p.Args["path"].(string)
					app := p.Source.(*model.App)
					data, err := app.LoadFile(path)
					if err != nil {
						return nil, err
					}
					return string(data), nil
				},
			},
			"rawAppConfig": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					app := p.Source.(*model.App)
					cfg, err := app.LoadAppConfigFile()
					if err != nil {
						return nil, err
					}
					return cfg, nil
				},
			},
			"rawSecretConfig": &graphql.Field{
				Type: graphql.NewNonNull(SecretConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					app := p.Source.(*model.App)
					cfg, err := app.LoadSecretConfigFile()
					if err != nil {
						return nil, err
					}
					return cfg, nil
				},
			},
			"effectiveAppConfig": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*model.App).Context.Config.AppConfig, nil
				},
			},
		},
	}),
	&model.App{},
	func(ctx context.Context, id string) (interface{}, error) {
		gqlCtx := GQLContext(ctx)
		lazy := gqlCtx.Apps.Get(id)
		return lazy.Value, nil
	},
)

var connApp = graphqlutil.NewConnectionDef(nodeApp)
