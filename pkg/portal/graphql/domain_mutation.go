package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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

			// Access Control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			domainModel, err := gqlCtx.DomainService.CreateCustomDomain(appID, domain)
			if err != nil {
				return nil, err
			}

			gqlCtx.Domains.Prime(domainModel.ID, domainModel)
			return graphqlutil.NewLazyValue(map[string]interface{}{
				"domain": gqlCtx.Domains.Load(domainModel.ID),
				"app":    gqlCtx.Apps.Load(appID),
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

			// Access Control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.DomainService.DeleteDomain(appID, domainID)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": gqlCtx.Apps.Load(appID),
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

			// Access Control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			domain, err := gqlCtx.DomainService.VerifyDomain(appID, domainID)
			if err != nil {
				return nil, err
			}

			gqlCtx.Domains.Prime(domain.ID, domain)
			return graphqlutil.NewLazyValue(map[string]interface{}{
				"domain": gqlCtx.Domains.Load(domain.ID),
				"app":    gqlCtx.Apps.Load(appID),
			}).Value, nil
		},
	},
)
