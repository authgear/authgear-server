package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

var createDomainInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateDomainInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
		"domain": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Domain name.",
		},
	},
})

var createDomainPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateDomainPayload",
	Fields: graphql.Fields{
		"domain": &graphql.Field{Type: graphql.NewNonNull(domain)},
		"app":    &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"createDomain",
	&graphql.Field{
		Description: "Create domain for target app",
		Type:        graphql.NewNonNull(createDomainPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createDomainInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			domain := input["domain"].(string)

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
					return gqlCtx.Domains.CreateDomain(appID, domain).
						Map(func(domain interface{}) (interface{}, error) {
							return map[string]interface{}{
								"domain": domain,
								"app":    app,
							}, nil
						}), nil
				}).Value, nil
		},
	},
)

var deleteDomainInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteDomainInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
		"domainID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Domain ID.",
		},
	},
})

var deleteDomainPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteDomainPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"deleteDomain",
	&graphql.Field{
		Description: "Delete domain of target app",
		Type:        graphql.NewNonNull(deleteDomainPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteDomainInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			domainID := input["domainID"].(string)

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
					return gqlCtx.Domains.DeleteDomain(appID, domainID).
						Map(func(domain interface{}) (interface{}, error) {
							return map[string]interface{}{
								"app": app,
							}, nil
						}), nil
				}).Value, nil
		},
	},
)

var verifyDomainInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "VerifyDomainInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
		"domainID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Domain ID.",
		},
	},
})

var verifyDomainPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "VerifyDomainPayload",
	Fields: graphql.Fields{
		"domain": &graphql.Field{Type: graphql.NewNonNull(domain)},
		"app":    &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"verifyDomain",
	&graphql.Field{
		Description: "Request verification of a domain of target app",
		Type:        graphql.NewNonNull(verifyDomainPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(verifyDomainInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			domainID := input["domainID"].(string)

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
					return gqlCtx.Domains.VerifyDomain(appID, domainID).
						Map(func(domain interface{}) (interface{}, error) {
							return map[string]interface{}{
								"domain": domain,
								"app":    app,
							}, nil
						}), nil
				}).Value, nil
		},
	},
)
