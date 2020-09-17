package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
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
			"rawAppConfig": &graphql.Field{
				Type: graphql.NewNonNull(AppConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					app := p.Source.(*model.App)
					data, err := app.LoadFile(configsource.AuthgearYAML)
					if err != nil {
						return nil, err
					}
					var jsonData interface{}
					if err := yaml.Unmarshal(data, &jsonData); err != nil {
						return nil, err
					}
					return jsonData, nil
				},
			},
			"rawSecretConfig": &graphql.Field{
				Type: graphql.NewNonNull(SecretConfig),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					app := p.Source.(*model.App)
					data, err := app.LoadFile(configsource.AuthgearSecretYAML)
					if err != nil {
						return nil, err
					}
					var jsonData interface{}
					if err := yaml.Unmarshal(data, &jsonData); err != nil {
						return nil, err
					}
					return jsonData, nil
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

var appConfigFile = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "AppConfigFile",
	Description: "A configuration file to update/create.",
	Fields: graphql.InputObjectConfigFieldMap{
		"path": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Path of the file.",
		},
		"content": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "New content of the file.",
		},
	},
})

var updateAppConfigInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateAppConfigInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to update.",
		},
		"updateFiles": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(appConfigFile)),
			Description: "Configuration files to update/create.",
		},
		"deleteFiles": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.String)),
			Description: "Path to configuration files to delete.",
		},
	},
})

var _ = registerMutationField(
	"updateAppConfig",
	&graphql.Field{
		Description: "Update app configuration files",
		Type:        graphql.NewNonNull(nodeApp),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateAppConfigInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			updateFiles, _ := input["updateFiles"].([]interface{})
			deleteFiles, _ := input["deleteFiles"].([]interface{})

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)
			lazy := gqlCtx.Apps.Get(appID)
			return lazy.
				Map(func(value interface{}) (interface{}, error) {
					app := value.(*model.App)
					var updateConfigFiles []*model.AppConfigFile
					var deleteConfigFiles []string
					for _, f := range updateFiles {
						f := f.(map[string]interface{})
						path := f["path"].(string)
						content := f["content"].(string)
						updateConfigFiles = append(updateConfigFiles, &model.AppConfigFile{
							Path:    path,
							Content: content,
						})
					}
					for _, p := range deleteFiles {
						deleteConfigFiles = append(deleteConfigFiles, p.(string))
					}

					return gqlCtx.Apps.UpdateConfig(app, updateConfigFiles, deleteConfigFiles), nil
				}).Value, nil
		},
	},
)
