package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

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

			err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

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

var createAppInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateAppInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "ID of the new app.",
		},
	},
})

var createAppPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateAppPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{
			Type: graphql.NewNonNull(nodeApp),
		},
	},
})

var _ = registerMutationField(
	"createApp",
	&graphql.Field{
		Description: "Create new app",
		Type:        graphql.NewNonNull(createAppPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createAppInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appID := input["id"].(string)

			gqlCtx := GQLContext(p.Context)
			viewer := gqlCtx.Viewer.Get()
			app := viewer.Map(func(u interface{}) (interface{}, error) {
				userID := u.(*model.User).ID
				return gqlCtx.Apps.Create(userID, appID), nil
			})
			return app.Map(func(app interface{}) (interface{}, error) {
				return map[string]interface{}{
					"app": app,
				}, nil
			}).Value, nil
		},
	},
)
