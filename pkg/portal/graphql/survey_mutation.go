package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

var saveOnboardingSurveyInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SaveOnboardingSurveyInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"surveyJSON": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Onboarding survey result JSON.",
		},
	},
})

var _ = registerMutationField(
	"saveOnboardingSurvey",
	&graphql.Field{
		Description: "Updates the current user's custom attribute with 'survey' key",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(saveOnboardingSurveyInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			surveyJSON := input["surveyJSON"].(string)
			gqlCtx := GQLContext(p.Context)

			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can fill onboarding survey")
			}
			actorID := sessionInfo.UserID

			entry := model.OnboardEntry{
				SurveyJSON: surveyJSON,
			}
			err := gqlCtx.OnboardService.SubmitOnboardEntry(
				entry,
				actorID,
			)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	},
)
