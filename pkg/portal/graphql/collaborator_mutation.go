package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

var deleteCollaboratorInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteCollaboratorInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"collaboratorID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Collaborator ID.",
		},
	},
})

var deleteCollaboratorPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteCollaboratorPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"deleteCollaborator",
	&graphql.Field{
		Description: "Delete collaborator of target app",
		Type:        graphql.NewNonNull(deleteCollaboratorPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteCollaboratorInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			collaboratorID := input["collaboratorID"].(string)

			gqlCtx := GQLContext(p.Context)
			return gqlCtx.Collaborators.DeleteCollaborator(collaboratorID).
				Map(func(value interface{}) (interface{}, error) {
					c := value.(*model.Collaborator)
					app := gqlCtx.Apps.Get(c.AppID)
					return map[string]interface{}{
						"app": app,
					}, nil
				}).Value, nil
		},
	},
)
