package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var tokenMutationLogger = slogutil.NewLogger("graphql-token-mutation")

var generateShortLivedAdminAPITokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "generateShortLivedAdminAPITokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to generate token for.",
		},
		"appSecretVisitToken": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "App secret visit token.",
		},
	},
})

var generateShortLivedAdminAPITokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "GenerateShortLivedAdminAPITokenPayload",
	Fields: graphql.Fields{
		"token": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var _ = registerMutationField(
	"generateShortLivedAdminAPIToken",
	&graphql.Field{
		Description: "Generate short-lived admin API token",
		Type:        generateShortLivedAdminAPITokenPayload,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(generateShortLivedAdminAPITokenInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(ctx)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can visit secrets")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			appSecretVisitToken := input["appSecretVisitToken"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(ctx, appID)
			if err != nil {
				return nil, err
			}

			secretConfig, _, err := gqlCtx.AppService.LoadAppSecretConfig(ctx, app, sessionInfo, appSecretVisitToken)
			if err != nil {
				return nil, err
			}

			if len(secretConfig.AdminAPISecrets) == 0 {
				panic(fmt.Sprintf("no admin API secret found for app: %s", appID))
			}

			keyID := secretConfig.AdminAPISecrets[0].KeyID
			privateKeyPEM := secretConfig.AdminAPISecrets[0].PrivateKeyPEM
			if privateKeyPEM == nil {
				return nil, apierrors.NewForbidden("invalid secret token")
			}

			logger := tokenMutationLogger.GetLogger(ctx)

			token, err := gqlCtx.TokenService.GenerateShortLivedAdminAPIToken(appID, keyID, *privateKeyPEM)
			if err != nil {
				logger.WithError(err).Error(ctx, "failed to generate short-lived admin API token")
				return nil, apierrors.NewInternalError("failed to generate short-lived admin API token")
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"token": token,
			}).Value, nil
		},
	},
)
