package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var createScopeSchema *validation.SimpleSchema

func init() {
	createScopeSchemaBuilder := validation.SchemaBuilder{}
	createScopeSchemaBuilder.Properties().Property("scope", (func() validation.SchemaBuilder {
		scopeBuilder := validation.SchemaBuilder{}
		scopeBuilder.Type(validation.TypeString).Format("x_scope_token")
		scopeBuilder.MinLength(1)
		scopeBuilder.MaxLength(100)
		scopeBuilder.Not((func() validation.SchemaBuilder {
			notBuilder := validation.SchemaBuilder{}
			scopesIntf := make([]interface{}, len(oauth.AllowedScopes))
			for i, s := range oauth.AllowedScopes {
				scopesIntf[i] = s
			}
			notBuilder.Enum(scopesIntf...)
			return notBuilder
		})())
		return scopeBuilder
	})())
	createScopeSchema = createScopeSchemaBuilder.ToSimpleSchema()
}

var createScopeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateScopeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"scope": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The scope string.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional description of the scope.",
		},
	},
})

var createScopePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateScopePayload",
	Fields: graphql.Fields{
		"scope": &graphql.Field{
			Type: graphql.NewNonNull(nodeScope),
		},
	},
})

var _ = registerMutationField(
	"createScope",
	&graphql.Field{
		Description: "Create a new scope.",
		Type:        graphql.NewNonNull(createScopePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createScopeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			ctx := p.Context
			input := p.Args["input"].(map[string]interface{})

			err := createScopeSchema.Validator().ValidateValue(ctx, input)
			if err != nil {
				return nil, err
			}

			resourceURI := input["resourceURI"].(string)
			scopeStr := input["scope"].(string)

			var description *string
			if str, ok := input["description"].(string); ok && str != "" {
				description = &str
			}

			options := &resourcescope.NewScopeOptions{
				ResourceURI: resourceURI,
				Scope:       scopeStr,
				Description: description,
			}

			gqlCtx := GQLContext(ctx)
			scope, err := gqlCtx.ResourceScopeFacade.CreateScope(ctx, options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateScopeExecutedEventPayload{
				Scope: *scope,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"scope": scope,
			}).Value, nil
		},
	},
)

var updateScopeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateScopeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"scope": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The scope string.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new description of the scope. Pass null if you do not need to update the description. Pass an empty string to remove the description.",
		},
	},
})

var updateScopePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateScopePayload",
	Fields: graphql.Fields{
		"scope": &graphql.Field{
			Type: graphql.NewNonNull(nodeScope),
		},
	},
})

var _ = registerMutationField(
	"updateScope",
	&graphql.Field{
		Description: "Update an existing scope.",
		Type:        graphql.NewNonNull(updateScopePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateScopeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceURI := input["resourceURI"].(string)
			scopeStr := input["scope"].(string)

			var newDescription *string
			if str, ok := input["description"].(string); ok {
				newDescription = &str
			}

			options := &resourcescope.UpdateScopeOptions{
				ResourceURI: resourceURI,
				Scope:       scopeStr,
				NewDesc:     newDescription,
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			originalScope, err := gqlCtx.ResourceScopeFacade.GetScope(ctx, resourceURI, scopeStr)
			if err != nil {
				return nil, err
			}

			scope, err := gqlCtx.ResourceScopeFacade.UpdateScope(ctx, options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationUpdateScopeExecutedEventPayload{
				OriginalScope: *originalScope,
				NewScope:      *scope,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"scope": scope,
			}).Value, nil
		},
	},
)

var deleteScopeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteScopeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"scope": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The scope string.",
		},
	},
})

var deleteScopePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteScopePayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"deleteScope",
	&graphql.Field{
		Description: "Delete a scope.",
		Type:        graphql.NewNonNull(deleteScopePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteScopeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceURI := input["resourceURI"].(string)
			scopeStr := input["scope"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			scope, err := gqlCtx.ResourceScopeFacade.GetScope(ctx, resourceURI, scopeStr)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.ResourceScopeFacade.DeleteScope(ctx, resourceURI, scopeStr)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteScopeExecutedEventPayload{
				Scope: *scope,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"ok": true,
			}, nil
		},
	},
)
