package graphql

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeOAuthClient = "OAuthClient"

var oauthClientSourceType = graphql.NewEnum(graphql.EnumConfig{
	Name: "OAuthClientSource",
	Values: graphql.EnumValueConfigMap{
		"STATIC": &graphql.EnumValueConfig{Value: string(model.OAuthClientSourceStatic)},
		"DCR":    &graphql.EnumValueConfig{Value: string(model.OAuthClientSourceDCR)},
		"CIMD":   &graphql.EnumValueConfig{Value: string(model.OAuthClientSourceCIMD)},
	},
})

// ErrDynamicClientsSourceStatic is returned when the dynamicClients query is
// filtered by STATIC. The enum is shared with OAuthClient.source, where STATIC
// is a real value, so it cannot be dropped from the type -- the argument has to
// reject it instead.
var ErrDynamicClientsSourceStatic = apierrors.NewInvalid("source: STATIC is not a dynamic client source")

// ParseDynamicClientsSourceArg reads the optional "source" argument of the
// dynamicClients query.
//
// STATIC is rejected rather than answered with an empty page. A statically
// configured client lives in authgear.yaml and is never persisted as a dynamic
// client, so `source: STATIC` can only ever be a caller mistake -- and an empty
// connection would read as "this project has no static clients", which is a
// different and possibly false claim. Failing loudly is the only answer that
// cannot be misread.
func ParseDynamicClientsSourceArg(args map[string]any) (*model.OAuthClientSource, error) {
	raw, ok := args["source"].(string)
	if !ok {
		// Absent (or null): no filter, list every dynamic client.
		return nil, nil
	}
	source := model.OAuthClientSource(raw)
	if source == model.OAuthClientSourceStatic {
		return nil, ErrDynamicClientsSourceStatic
	}
	return &source, nil
}

// MaxDynamicClientsClientIDs bounds the clientIDs argument of the
// dynamicClients query. It is also that query's page size limit while the
// argument is set, so one page always holds every match.
const MaxDynamicClientsClientIDs = 1000

var ErrDynamicClientsTooManyClientIDs = apierrors.NewInvalid(fmt.Sprintf("clientIDs: at most %d client IDs", MaxDynamicClientsClientIDs))

// ParseDynamicClientsClientIDsArg reads the optional "clientIDs" argument of
// the dynamicClients query. Absent or null means no filter (nil); an empty
// list is a filter that matches nothing (non-nil, empty).
func ParseDynamicClientsClientIDsArg(args map[string]any) ([]string, error) {
	arr, ok := args["clientIDs"].([]any)
	if !ok {
		return nil, nil
	}
	if len(arr) > MaxDynamicClientsClientIDs {
		return nil, ErrDynamicClientsTooManyClientIDs
	}
	clientIDs := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			clientIDs = append(clientIDs, s)
		}
	}
	return clientIDs, nil
}

var oauthClientKindType = graphql.NewEnum(graphql.EnumConfig{
	Name: "OAuthClientKind",
	Values: graphql.EnumValueConfigMap{
		"FIRST_PARTY": &graphql.EnumValueConfig{Value: string(model.OAuthClientKindFirstParty)},
		"THIRD_PARTY": &graphql.EnumValueConfig{Value: string(model.OAuthClientKindThirdParty)},
	},
})

// authenticationFlowAllowlistFlowType/GroupType/Type are a minimal
// placeholder shape (AuthenticationFlowAllowlist { groups { name } flows {
// type name } }) with no backing Go data yet — every OAuthClient this part
// can produce is DCR-sourced, and client.md fixes authenticationFlowAllowlist
// at null for DCR clients. Defining the shape now keeps the schema
// conformant with client.md rather than omitting the field.
var authenticationFlowAllowlistFlowType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthenticationFlowAllowlistFlow",
	Fields: graphql.Fields{
		"type": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"name": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var authenticationFlowAllowlistGroupType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthenticationFlowAllowlistGroup",
	Fields: graphql.Fields{
		"name": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var authenticationFlowAllowlistType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthenticationFlowAllowlist",
	Fields: graphql.Fields{
		"groups": &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(authenticationFlowAllowlistGroupType)))},
		"flows":  &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(authenticationFlowAllowlistFlowType)))},
	},
})

var nodeOAuthClient = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeOAuthClient,
		Description: "A client that exists outside authgear.yaml: DCR-registered or CIMD-resolved.",
		Interfaces: []*graphql.Interface{
			// Uses clientID as its external key, but the Node id is still
			// the relay global id derived from the row uuid — see
			// pkg/admin/loader/oauth_client.go.
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": entityIDField(typeOAuthClient),
			"clientID": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"source": &graphql.Field{
				Type: graphql.NewNonNull(oauthClientSourceType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					c := p.Source.(*model.OAuthClient)
					return string(c.Source), nil
				},
			},
			"kind": &graphql.Field{
				Type: graphql.NewNonNull(oauthClientKindType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					c := p.Source.(*model.OAuthClient)
					return string(c.Kind), nil
				},
			},
			"isConfidential":                         &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"isServiceClient":                        &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"applicationType":                        &graphql.Field{Type: graphql.String},
			"name":                                   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"clientName":                             &graphql.Field{Type: graphql.String},
			"clientURI":                              &graphql.Field{Type: graphql.String},
			"logoURI":                                &graphql.Field{Type: graphql.String},
			"tosURI":                                 &graphql.Field{Type: graphql.String},
			"policyURI":                              &graphql.Field{Type: graphql.String},
			"redirectURIs":                           &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"postLogoutRedirectURIs":                 &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"grantTypes":                             &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"responseTypes":                          &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"accessTokenLifetimeSeconds":             &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"refreshTokenLifetimeSeconds":            &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"refreshTokenIdleTimeoutEnabled":         &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"refreshTokenIdleTimeoutSeconds":         &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"refreshTokenRotationEnabled":            &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"issueJWTAccessToken":                    &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"maxConcurrentSession":                   &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"customUIURI":                            &graphql.Field{Type: graphql.String},
			"app2appEnabled":                         &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"app2appInsecureDeviceKeyBindingEnabled": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"dpopDisabled":                           &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"authenticationFlowAllowlist": &graphql.Field{
				Type: authenticationFlowAllowlistType,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					// Always null for DCR clients (client.md); no source
					// carries this data yet.
					return nil, nil
				},
			},
			"preAuthenticatedURLEnabled":        &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"preAuthenticatedURLAllowedOrigins": &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"replaceProjectLogoWithLogoURI":     &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"registeredAt":                      &graphql.Field{Type: graphql.DateTime},
			"lastFetchedAt":                     &graphql.Field{Type: graphql.DateTime},
		},
	}),
	&model.OAuthClient{},
	func(ctx context.Context, gqlCtx *Context, id string) (any, error) {
		return gqlCtx.DynamicClients.Load(ctx, id).Value, nil
	},
)

var connDynamicClient = graphqlutil.NewConnectionDef(nodeOAuthClient)
