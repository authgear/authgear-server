package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/lib/analytic"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type OnboardServiceAdminAPIService interface {
	SelfDirector(ctx context.Context, actorUserID string, usage Usage) (func(*http.Request), error)
}

type OnboardServicePosthogService interface {
	Batch(ctx context.Context, events []json.RawMessage) error
}

type OnboardService struct {
	HTTPClient     HTTPClient
	AuthgearConfig *portalconfig.AuthgearConfig
	AdminAPI       OnboardServiceAdminAPIService

	PosthogService OnboardServicePosthogService
}

func (s *OnboardService) graphqlDo(ctx context.Context, params graphqlutil.DoParams, actorID string) (*graphql.Result, error) {
	r, err := http.NewRequestWithContext(ctx, "POST", "/graphql", nil)
	if err != nil {
		return nil, err
	}

	director, err := s.AdminAPI.SelfDirector(ctx, actorID, UsageInternal)
	if err != nil {
		return nil, err
	}
	director(r)

	result, err := graphqlutil.HTTPDo(s.HTTPClient.Client, r, params)
	if err != nil {
		return nil, err
	}

	if result.HasErrors() {
		return nil, fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}
	return result, nil
}

func (s *OnboardService) SubmitOnboardEntry(ctx context.Context, entry model.OnboardEntry, actorID string) error {
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

	_, err := s.graphqlDo(ctx, params, actorID)
	if err != nil {
		return err
	}

	event := map[string]interface{}{
		"event":       "onboardingSurvey.set-completed-survey",
		"distinct_id": actorID,
		"properties": map[string]interface{}{
			"$set": map[string]interface{}{
				"survey_json": json.RawMessage(entry.SurveyJSON),
			},
		},
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal survey entry event: %w", err)
	}
	err = s.PosthogService.Batch(ctx, []json.RawMessage{eventBytes})
	if err != nil && !errors.Is(err, analytic.ErrMissingPosthogCredential) {
		return fmt.Errorf("failed to send survey entry event: %w", err)
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

func (s *OnboardService) CheckOnboardingSurveyCompletion(ctx context.Context, actorID string) (bool, error) {
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

	result, err := s.graphqlDo(ctx, params, actorID)
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
