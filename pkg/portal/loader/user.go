package loader

import (
	"fmt"
	"net/http"

	"github.com/authgear/graphql-go-relay"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserLoaderAdminAPIService interface {
	SelfDirector() (func(*http.Request), error)
}

type UserLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	AdminAPI UserLoaderAdminAPIService
}

func NewUserLoader(adminAPI UserLoaderAdminAPIService) *UserLoader {
	l := &UserLoader{
		AdminAPI: adminAPI,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *UserLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	var ids []string
	for _, iface := range keys {
		key := iface.(string)
		ids = append(ids, relay.ToGlobalID("User", key))
	}

	params := graphqlutil.DoParams{
		OperationName: "getUserNodes",
		Query: `
		query getUserNodes($ids: [ID!]!) {
			nodes(ids: $ids) {
				... on User {
					id
					verifiedClaims {
						name
						value
					}
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"ids": ids,
		},
	}

	r, err := http.NewRequest("POST", "/graphql", nil)
	if err != nil {
		return nil, err
	}

	director, err := l.AdminAPI.SelfDirector()
	if err != nil {
		return nil, err
	}

	director(r)

	result, err := graphqlutil.HTTPDo(r, params)
	if err != nil {
		return nil, err
	}

	if result.HasErrors() {
		return nil, fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}

	var userModels []interface{}

	data := result.Data.(map[string]interface{})
	nodes := data["nodes"].([]interface{})
	for _, iface := range nodes {
		// It could be null.
		userNode, ok := iface.(map[string]interface{})
		if !ok {
			userModels = append(userModels, nil)
		} else {
			userModel := &model.User{}
			globalID := userNode["id"].(string)
			userModel.ID = globalID

			// Use the last email claim.
			verifiedClaims := userNode["verifiedClaims"].([]interface{})
			for _, iface := range verifiedClaims {
				claim := iface.(map[string]interface{})
				name := claim["name"].(string)
				value := claim["value"].(string)
				if name == "email" {
					userModel.Email = value
				}
			}

			userModels = append(userModels, userModel)
		}
	}

	return userModels, nil
}
