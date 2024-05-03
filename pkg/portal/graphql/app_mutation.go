package graphql

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
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
		"checksum": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The checksum of the original resource file. If provided, it will be used to detect conflict.",
		},
	},
})

var oauthSSOProviderClientSecretInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "OAuthSSOProviderClientSecretInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"originalAlias": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"newAlias": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"newClientSecret": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
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

var oauthClientSecretsGenerateDataInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "OAuthClientSecretsGenerateDataInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"clientID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var oauthClientSecretsCleanupDataInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "OAuthClientSecretsCleanupDataInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"keepClientIDs": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
		},
	},
})

var adminAPIAuthKeyDeleteDataInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AdminAPIAuthKeyDeleteDataInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"keyID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var smtpSecretUpdateInstructionsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SmtpSecretUpdateInstructionsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"action": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"data": &graphql.InputObjectFieldConfig{
			Type: smtpSecretInput,
		},
	},
})

var oauthSSOProviderClientSecretsUpdateInstructionsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "OAuthSSOProviderClientSecretsUpdateInstructionsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"action": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"data": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.NewNonNull(oauthSSOProviderClientSecretInput)),
		},
	},
})

var oauthClientSecretsUpdateInstructionsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "OAuthClientSecretsUpdateInstructionsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"action": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"generateData": &graphql.InputObjectFieldConfig{
			Type: oauthClientSecretsGenerateDataInput,
		},
		"cleanupData": &graphql.InputObjectFieldConfig{
			Type: oauthClientSecretsCleanupDataInput,
		},
	},
})

var adminAPIAuthKeyUpdateInstructionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AdminAPIAuthKeyUpdateInstructionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"action": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"deleteData": &graphql.InputObjectFieldConfig{
			Type: adminAPIAuthKeyDeleteDataInput,
		},
	},
})

var secretConfigUpdateInstructionsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SecretConfigUpdateInstructionsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"oauthSSOProviderClientSecrets": &graphql.InputObjectFieldConfig{
			Type: oauthSSOProviderClientSecretsUpdateInstructionsInput,
		},
		"smtpSecret": &graphql.InputObjectFieldConfig{
			Type: smtpSecretUpdateInstructionsInput,
		},
		"oauthClientSecrets": &graphql.InputObjectFieldConfig{
			Type: oauthClientSecretsUpdateInstructionsInput,
		},
		"adminAPIAuthKey": &graphql.InputObjectFieldConfig{
			Type: adminAPIAuthKeyUpdateInstructionInput,
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
		"appConfigChecksum": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The checksum of appConfig. If provided, it will be used to detect conflict.",
		},
		"secretConfigUpdateInstructions": &graphql.InputObjectFieldConfig{
			Type:        secretConfigUpdateInstructionsInput,
			Description: "update secret config instructions.",
		},
		"secretConfigUpdateInstructionsChecksum": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The checksum of secretConfig. If provided, it will be used to detect conflict.",
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
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can update app")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			updates, _ := input["updates"].([]interface{})
			appConfigJSONValue := input["appConfig"]
			appConfigChecksum, _ := input["appConfigChecksum"].(string)
			secretConfigUpdateInstructionsJSONValue := input["secretConfigUpdateInstructions"]
			secretConfigUpdateInstructionsChecksum, _ := input["secretConfigUpdateInstructionsChecksum"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
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

			originalAppConfig := app.Context.Config.AppConfig

			var resourceUpdates []appresource.Update
			var auditedUpdatePaths []string = []string{}
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

				checksum, _ := f["checksum"].(string)

				resourceUpdates = append(resourceUpdates, appresource.Update{
					Path:     path,
					Data:     data,
					Checksum: checksum,
				})
				auditedUpdatePaths = append(auditedUpdatePaths, path)
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
					Path:     configsource.AuthgearYAML,
					Data:     appConfigYAML,
					Checksum: appConfigChecksum,
				})
			}

			// Update authgear.secrets.yaml
			if secretConfigUpdateInstructionsJSONValue != nil {
				secretConfigUpdateInstructionsJSON, err := json.Marshal(secretConfigUpdateInstructionsJSONValue)
				if err != nil {
					return nil, err
				}

				resourceUpdates = append(resourceUpdates, appresource.Update{
					Path:     configsource.AuthgearSecretYAML,
					Data:     secretConfigUpdateInstructionsJSON,
					Checksum: secretConfigUpdateInstructionsChecksum,
				})
			}

			err = gqlCtx.AppService.UpdateResources(app, resourceUpdates)
			if err != nil {
				return nil, err
			}

			newApp, err := gqlCtx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			newAppConfig := newApp.Context.Config.AppConfig

			appConfigDiff, err := config.DiffAppConfig(originalAppConfig, newAppConfig)
			if err != nil {
				return nil, err
			}
			var updatedSecrets []string = []string{}
			if secretUpdateInstructionsMap, ok := secretConfigUpdateInstructionsJSONValue.(map[string]interface{}); ok {
				for k := range secretUpdateInstructionsMap {
					updatedSecrets = append(updatedSecrets, k)
				}
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectAppUpdatedEventPayload{
				AppConfigOld:     originalAppConfig,
				AppConfigNew:     newAppConfig,
				AppConfigDiff:    appConfigDiff,
				UpdatedSecrets:   updatedSecrets,
				UpdatedResources: auditedUpdatePaths,
			})
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
		"phoneNumber": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Phone number of the new app.",
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
			phoneNumber := input["phoneNumber"].(string)

			gqlCtx := GQLContext(p.Context)

			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can create app")
			}

			actorID := sessionInfo.UserID

			err := checkAppQuota(gqlCtx, actorID)
			if err != nil {
				return nil, err
			}

			if phoneNumber != "" {
				entry := model.OnboardEntry{
					PhoneNumber: phoneNumber,
				}
				err := gqlCtx.OnboardService.SubmitOnboardEntry(
					entry,
					actorID,
				)
				if err != nil {
					return nil, err
				}
			}

			app, err := gqlCtx.AppService.Create(actorID, appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectAppCreatedEventPayload{
				AppConfig: app.Context.Config.AppConfig,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": gqlCtx.Apps.Load(appID),
			}).Value, nil
		},
	},
)

var skipAppTutorialInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SkipAppTutorialInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "ID of the app.",
		},
	},
})

var skipAppTutorialPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SkipAppTutorialPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{
			Type: graphql.NewNonNull(nodeApp),
		},
	},
})

var _ = registerMutationField(
	"skipAppTutorial",
	&graphql.Field{
		Description: "Skip the tutorial of the app",
		Type:        graphql.NewNonNull(skipAppTutorialPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(skipAppTutorialInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can create app")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["id"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.TutorialService.Skip(appID)
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

var skipAppTutorialProgressInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SkipAppTutorialProgressInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "ID of the app.",
		},
		"progress": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The progress to skip.",
		},
	},
})

var skipAppTutorialProgressPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SkipAppTutorialProgressPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{
			Type: graphql.NewNonNull(nodeApp),
		},
	},
})

var _ = registerMutationField(
	"skipAppTutorialProgress",
	&graphql.Field{
		Description: "Skip a progress of the tutorial of the app",
		Type:        graphql.NewNonNull(skipAppTutorialProgressPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(skipAppTutorialProgressInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can create app")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["id"].(string)
			progressStr := input["progress"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			progress, ok := tutorial.ProgressFromString(progressStr)
			if !ok {
				return nil, apierrors.NewInvalid("invalid progress")
			}

			err = gqlCtx.TutorialService.RecordProgresses(appID, []tutorial.Progress{progress})
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

var generateAppSecretVisitTokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "GenerateAppSecretVisitTokenPayloadPayload",
	Fields: graphql.Fields{
		"token": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The generated token",
		},
	},
})

var generateAppSecretVisitTokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "GenerateAppSecretVisitTokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "ID of the app.",
		},
		"secrets": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(appSecretKey))),
			Description: "Secrets to visit.",
		},
	},
})

var _ = registerMutationField(
	"generateAppSecretVisitToken",
	&graphql.Field{
		Description: "Generate a token for visiting app secrets",
		Type:        graphql.NewNonNull(generateAppSecretVisitTokenPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(generateAppSecretVisitTokenInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can visit secrets")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["id"].(string)
			secretsRaw := input["secrets"].([]interface{})
			var appSecretKeys []string = []string{}
			var secrets []config.SecretKey = []config.SecretKey{}
			for _, s := range secretsRaw {
				appSecretKey := s.(AppSecretKey)
				appSecretKeys = append(appSecretKeys, string(appSecretKey))
				configSecretKey := secretKeyToConfigKeyMap[appSecretKey]
				secrets = append(secrets, configSecretKey)
			}

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
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

			token, err := gqlCtx.AppService.GenerateSecretVisitToken(app, sessionInfo, secrets)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectAppSecretViewedEventPayload{
				Secrets: appSecretKeys,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"token": token.TokenID,
			}).Value, nil
		},
	},
)

var generateTestTokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "GenerateTestTokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "ID of the app.",
		},
		"returnUri": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "URI to return to in the tester page",
		},
	},
})

var generateTestTokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "GenerateTestTokenPayload",
	Fields: graphql.Fields{
		"token": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The generated token",
		},
	},
})

var _ = registerMutationField(
	"generateTesterToken",
	&graphql.Field{
		Description: "Generate a token for tester",
		Type:        graphql.NewNonNull(generateTestTokenPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(generateTestTokenInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can generate tester token")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["id"].(string)
			returnURI := input["returnUri"].(string)
			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
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

			token, err := gqlCtx.AppService.GenerateTesterToken(app, returnURI)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"token": token.TokenID,
			}).Value, nil
		},
	},
)

func checkAppQuota(ctx *Context, userID string) error {
	quota, err := ctx.AppService.GetProjectQuota(userID)
	if err != nil {
		return err
	}

	if quota < 0 {
		// Negative quota: skip checking
		return nil
	}

	numOwnedApps, err := ctx.CollaboratorService.GetProjectOwnerCount(userID)
	if err != nil {
		return err
	}

	if numOwnedApps >= quota {
		return QuotaExceeded.Errorf("you can only own a maximum of %d apps", quota)
	}

	return nil
}
