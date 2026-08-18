package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
)

var accessPolicyType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AccessPolicy",
	Fields: graphql.Fields{
		"allowDynamicThirdPartyClientAccess": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether a dynamically registered (DCR) or static third-party client can request this resource/scope via the resource parameter.",
		},
	},
})

var accessPolicyInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AccessPolicyInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"allowDynamicThirdPartyClientAccess": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
	},
})

// decodeAccessPolicyInput returns nil when "accessPolicy" was omitted from
// input entirely, or explicitly passed as null (leave unchanged on update,
// default-false on create), and a non-nil *model.AccessPolicy otherwise.
func decodeAccessPolicyInput(input map[string]any) *model.AccessPolicy {
	raw, ok := input["accessPolicy"]
	if !ok || raw == nil {
		return nil
	}
	m, _ := raw.(map[string]any)
	ap := &model.AccessPolicy{}
	if v, ok := m["allowDynamicThirdPartyClientAccess"].(bool); ok {
		ap.AllowDynamicThirdPartyClientAccess = v
	}
	return ap
}
