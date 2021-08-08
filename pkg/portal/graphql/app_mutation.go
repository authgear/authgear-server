package graphql

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var appResourceUpdate = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "AppResourceUpdate",
	Description: "Update to resource file.",
	Fields: graphql.InputObjectConfigFieldMap{
		"path": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Path of the resource file to update.",
		},
		"data": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "New data of the resource file. Set to null to remove it.",
		},
	},
})

var oauthClientSecretInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "OauthClientSecretInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"alias": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"clientSecret": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var smtpSecretInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMTPSecretInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"host": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"port": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"username": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"password": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var secretConfigInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SecretConfigInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"oauthClientSecrets": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.NewNonNull(oauthClientSecretInput)),
		},
		"smtpSecret": &graphql.InputObjectFieldConfig{
			Type: smtpSecretInput,
		},
	},
})

var updateAppInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateAppInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to update.",
		},
		"updates": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(appResourceUpdate)),
			Description: "Resource file updates.",
		},
		"appConfig": &graphql.InputObjectFieldConfig{
			Type:        AppConfig,
			Description: "authgear.yaml in JSON.",
		},
		"secretConfig": &graphql.InputObjectFieldConfig{
			Type:        secretConfigInput,
			Description: "secrets to update.",
		},
	},
})

var updateAppPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateAppPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"updateApp",
	&graphql.Field{
		Description: "Update app",
		Type:        graphql.NewNonNull(updateAppPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateAppInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			updates, _ := input["updates"].([]interface{})
			appConfigJSONValue := input["appConfig"]
			secretConfigJSONValue := input["secretConfig"]

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			var resourceUpdates []appresource.Update
			for _, f := range updates {
				f := f.(map[string]interface{})
				path := f["path"].(string)
				if path == configsource.AuthgearYAML {
					return nil, errors.New("direct update on authgear.yaml is disallowed")
				}
				if path == configsource.AuthgearSecretYAML {
					return nil, errors.New("direct update on authgear.secrets.yaml is disallowed")
				}

				var data []byte
				if stringData, ok := f["data"].(string); ok {
					data, err = base64.StdEncoding.DecodeString(stringData)
					if err != nil {
						return nil, err
					}
				}

				resourceUpdates = append(resourceUpdates, appresource.Update{
					Path: path,
					Data: data,
				})
			}

			// Update authgear.yaml
			if appConfigJSONValue != nil {
				appConfigJSON, err := json.Marshal(appConfigJSONValue)
				if err != nil {
					return nil, err
				}
				appConfigYAML, err := yaml.JSONToYAML(appConfigJSON)
				if err != nil {
					return nil, err
				}

				resourceUpdates = append(resourceUpdates, appresource.Update{
					Path: configsource.AuthgearYAML,
					Data: appConfigYAML,
				})
			}

			// Update authgear.secrets.yaml
			if secretConfigJSONValue != nil {
				secretConfigJSON, err := json.Marshal(secretConfigJSONValue)
				if err != nil {
					return nil, err
				}

				var structured model.SecretConfig
				err = json.Unmarshal(secretConfigJSON, &structured)
				if err != nil {
					return nil, err
				}

				secretConfigYAML, err := structured.ToYAMLForUpdate(app.Context.Config.SecretConfig)
				if err != nil {
					return nil, err
				}

				resourceUpdates = append(resourceUpdates, appresource.Update{
					Path: configsource.AuthgearSecretYAML,
					Data: secretConfigYAML,
				})
			}

			err = gqlCtx.AppService.UpdateResources(app, resourceUpdates)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": gqlCtx.Apps.Load(appID),
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

			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, AccessDenied.New("only authenticated users can create app")
			}

			actorID := sessionInfo.UserID

			err := checkAppQuota(gqlCtx, actorID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AppService.Create(actorID, appID)
			if err != nil {
				return nil, err
			}

			appLazy := gqlCtx.Apps.Load(appID)
			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": appLazy,
			}).Value, nil
		},
	},
)

func checkAppQuota(ctx *Context, userID string) error {
	quota, err := ctx.AppService.GetMaxOwnedApps(userID)
	if err != nil {
		return err
	}

	if quota < 0 {
		// Negative quota: skip checking
		return nil
	}

	collaborators, err := ctx.CollaboratorService.ListCollaboratorsByUser(userID)
	if err != nil {
		return err
	}

	numOwnedApps := 0
	for _, c := range collaborators {
		if c.Role == model.CollaboratorRoleOwner {
			numOwnedApps++
		}
	}

	if numOwnedApps >= quota {
		return QuotaExceeded.Errorf("you can only own a maximum of %d apps", quota)
	}

	return nil
}
