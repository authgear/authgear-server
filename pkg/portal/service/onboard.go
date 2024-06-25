package service

import (
	"fmt"
	"net/http"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type OnboardServiceAdminAPIService interface {
	SelfDirector(actorUserID string, usage Usage) (func(*http.Request), error)
}

type OnboardService struct {
	AuthgearConfig *portalconfig.AuthgearConfig
	AdminAPI       OnboardServiceAdminAPIService
}

func (s *OnboardService) graphqlDo(params graphqlutil.DoParams, actorID string) (*graphql.Result, error) {
	r, err := http.NewRequest("POST", "/graphql", nil)
	if err != nil {
		return nil, err
	}

	director, err := s.AdminAPI.SelfDirector(actorID, UsageInternal)
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
	return result, nil
}

func (s *OnboardService) SubmitOnboardEntry(entry model.OnboardEntry, actorID string) error {
	id := relay.ToGlobalID("User", actorID)
	params := graphqlutil.DoParams{
		OperationName: "submitOnboardEntry",
		Query: `
		mutation submitOnboardEntry($userID: ID!, $customAttributes: UserCustomAttributes!) {
			updateUser(
				input: {userID: $userID, customAttributes: $customAttributes}
			) {
				user {
					id
					updatedAt
					customAttributes
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"userID":           id,
			"customAttributes": entry,
		},
	}

	_, err := s.graphqlDo(params, actorID)
	if err != nil {
		return err
	}
	return nil
}

func unwrap(thing interface{}, keys []string) (interface{}, bool) {
	if len(keys) == 0 {
		return thing, true
	}
	mapThing, ok := thing.(map[string]interface{})
	if !ok {
		return nil, false
	}
	value, ok := mapThing[keys[0]]
	if !ok {
		return nil, false
	}
	return unwrap(value, keys[1:])
}

func (s *OnboardService) CheckOnboardingSurveyCompletion(actorID string) (bool, error) {
	id := relay.ToGlobalID("User", actorID)
	params := graphqlutil.DoParams{
		OperationName: "checkOnboardEntry",
		Query: `
		query checkOnboardEntry($userID: ID!) {
			node(id: $userID) {
				... on User {
					customAttributes
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"userID": id,
		},
	}

	result, err := s.graphqlDo(params, actorID)
	if err != nil {
		return false, err
	}
	surveyCustAttrIface, ok := unwrap(result.Data, []string{"node", "customAttributes", "onboarding_survey_json"})
	surveyCustAttr, ok2 := surveyCustAttrIface.(string)
	if !ok || !ok2 || surveyCustAttr == "" {
		return false, nil
	}
	return true, nil
}
