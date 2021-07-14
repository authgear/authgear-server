package graphql

import (
	"encoding/base64"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

var updateAppResourcesInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateAppResourcesInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to update.",
		},
		"updates": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(appResourceUpdate)),
			Description: "Resource file updates.",
		},
	},
})

var updateAppResourcesPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateAppResourcesPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"updateAppResources",
	&graphql.Field{
		Description: "Update app resource files",
		Type:        graphql.NewNonNull(updateAppResourcesPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateAppResourcesInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			updates, _ := input["updates"].([]interface{})

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

			err = gqlCtx.AppService.UpdateResources(app, resourceUpdates)
			if err != nil {
				return nil, err
			}

			// App is not primed here intentionally.
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
