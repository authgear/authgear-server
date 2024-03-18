package service

import (
	"fmt"
	"net/http"

	relay "github.com/authgear/graphql-go-relay"

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

	r, err := http.NewRequest("POST", "/graphql", nil)
	if err != nil {
		return err
	}

	director, err := s.AdminAPI.SelfDirector(actorID, UsageInternal)
	if err != nil {
		return err
	}

	director(r)

	result, err := graphqlutil.HTTPDo(r, params)
	if err != nil {
		return err
	}

	if result.HasErrors() {
		return fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}

	return nil
}
