package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
)

var accessPolicyType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AccessPolicy",
	Fields: graphql.Fields{
		"allowStaticFirstPartyClientAccess": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "When true, any client declared in authgear.yaml with a first-party application type (spa, traditional_webapp, native, confidential) may access this Resource or Scope without a per-client association. Does not cover m2m clients, and has no effect on the client_credentials grant.",
		},
		"allowStaticThirdPartyClientAccess": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "When true, any third_party_app client declared in authgear.yaml may access this Resource or Scope without a per-client association.",
		},
		"allowDynamicFirstPartyClientAccess": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "When true, any dynamically registered first-party client (DCR-registered with a first-party Initial Access Token, or CIMD-resolved as first-party) may access this Resource or Scope without a per-client association.",
		},
		"allowDynamicThirdPartyClientAccess": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "When true, any dynamically registered third-party client (DCR/CIMD) may access this Resource or Scope without a per-client association. Static third-party clients are never covered by this flag.",
		},
	},
})

var accessPolicyInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AccessPolicyInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"allowStaticFirstPartyClientAccess": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Default false on create; unchanged on update.",
		},
		"allowStaticThirdPartyClientAccess": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Default false on create; unchanged on update.",
		},
		"allowDynamicFirstPartyClientAccess": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Default false on create; unchanged on update.",
		},
		"allowDynamicThirdPartyClientAccess": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Default false on create; unchanged on update.",
		},
	},
})

// decodeAccessPolicyInput returns nil when "accessPolicy" was omitted from
// input entirely, or explicitly passed as null (leave unchanged on update,
// default-false on create), and a non-nil *model.AccessPolicyPatch
// otherwise. Only the fields actually present in the GraphQL input are set
// on the patch, which is what makes the patch a merge rather than a
// replace: a field a caller never mentioned stays nil, and Store.Update*
// leaves the corresponding stored key alone.
func decodeAccessPolicyInput(input map[string]any) *model.AccessPolicyPatch {
	raw, ok := input["accessPolicy"]
	if !ok || raw == nil {
		return nil
	}
	m, _ := raw.(map[string]any)
	patch := &model.AccessPolicyPatch{}
	if v, ok := m["allowStaticFirstPartyClientAccess"].(bool); ok {
		patch.AllowStaticFirstPartyClientAccess = &v
	}
	if v, ok := m["allowStaticThirdPartyClientAccess"].(bool); ok {
		patch.AllowStaticThirdPartyClientAccess = &v
	}
	if v, ok := m["allowDynamicFirstPartyClientAccess"].(bool); ok {
		patch.AllowDynamicFirstPartyClientAccess = &v
	}
	if v, ok := m["allowDynamicThirdPartyClientAccess"].(bool); ok {
		patch.AllowDynamicThirdPartyClientAccess = &v
	}
	return patch
}
